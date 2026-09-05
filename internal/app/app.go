package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/moleship-org/moleship/internal/api/handler"
	"github.com/moleship-org/moleship/internal/api/middleware"
	"github.com/moleship-org/moleship/internal/config"
	"github.com/moleship-org/moleship/internal/domain/podman"
	"github.com/moleship-org/moleship/internal/domain/systemd"
	"github.com/moleship-org/moleship/internal/services/auth"
	"github.com/moleship-org/moleship/internal/services/quadlet"
)

type Application struct {
	cfg *Config

	router chi.Router

	server *http.Server

	// --- Adapters ---

	podmanAdapter  *podman.Podman
	systemdAdapter *systemd.Systemd

	// --- Services ---

	quadletFS *quadlet.Filesystem

	quadletSvc *quadlet.QuadletService

	authSvc *auth.AuthService
}

func New(opts ...Option) *Application {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	a := new(Application)
	a.cfg = cfg
	a.router = chi.NewRouter()

	return a
}

func (a *Application) Start(ctx context.Context) error {
	a.server = &http.Server{
		Addr:              a.Addr(),
		Handler:           a.router,
		ReadTimeout:       a.cfg.ReadTimeout,
		ReadHeaderTimeout: a.cfg.ReadHeaderTimeout,
		WriteTimeout:      a.cfg.WriteTimeout,
		IdleTimeout:       a.cfg.IdleTimeout,
		MaxHeaderBytes:    a.cfg.MaxHeaderBytes,
	}

	serverErrors := make(chan error, 1)
	go func() {
		a.Logger().Info(fmt.Sprintf("Application running on http://localhost%s/ - Press CTRL+C to exit", a.Addr()))
		serverErrors <- a.server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		a.Logger().Error(err.Error())

	case <-shutdown:
		a.Logger().Warn("Starting application shutdown...")

		ctx, cancel := context.WithTimeout(ctx, a.cfg.ShutdownTimeout)
		defer cancel()

		if shutErr := a.server.Shutdown(ctx); shutErr != nil {
			a.Logger().Error("Shutdown error", slog.String("error", shutErr.Error()))
			if closeErr := a.server.Close(); closeErr != nil {
				return errors.Join(closeErr, shutErr)
			}
			return shutErr
		}
	}

	return nil
}

func (a Application) Addr() string {
	return fmt.Sprintf(":%d", a.cfg.Port)
}

func (a *Application) Config() *Config {
	if a.cfg == nil {
		a.cfg = DefaultConfig()
		return a.cfg
	}
	return a.cfg
}

func (a *Application) Logger() *slog.Logger {
	if a.cfg.Logger == nil {
		return slog.Default()
	}
	return a.cfg.Logger
}

func (a *Application) MountRoutes() error {
	if a.podmanAdapter == nil {
		return errors.New("invalid podman adapter: nil")
	}
	if a.systemdAdapter == nil {
		return errors.New("invalid systemd adapter: nil")
	}
	if a.quadletSvc == nil {
		return errors.New("invalid quadlet service: nil")
	}
	if a.authSvc == nil {
		return errors.New("invalid auth service: nil")
	}

	a.router.Use(middleware.ContextInjector(a.Logger()))
	a.router.Use(middleware.RequestLogger())

	a.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.ALLOWED_ORIGINS,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: config.IsReleaseMode(),
		MaxAge:           300,
	}))

	a.router.Use(chi_middleware.Recoverer)
	a.router.Use(chi_middleware.RequestID)
	a.router.Use(chi_middleware.CleanPath)

	healthHandler := handler.NewHealth(a.Logger())
	authHandler := handler.NewAuth(a.Logger(), a.authSvc)
	libpodHandler := handler.NewLibpod(a.Logger(), a.podmanAdapter)
	systemdHandler := handler.NewSystemd(a.Logger(), a.systemdAdapter)
	quadletHandler := handler.NewQuadlet(a.Logger(), a.quadletSvc)

	publicRate := config.PUBLIC_RATE_LIMIT
	publicLimiter := tollbooth.NewLimiter(publicRate, &limiter.ExpirableOptions{
		DefaultExpirationTTL: 1 * time.Hour,
	})
	publicLimiter.SetBurst(config.PUBLIC_BURST_LIMIT)

	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(tollbooth.HTTPMiddleware(publicLimiter))

				healthHandler.Mount(r)
				authHandler.Mount(r)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth())

				libpodHandler.Mount(r)
				systemdHandler.Mount(r)
				quadletHandler.Mount(r)
			})
		})
	})

	return nil
}

func (a *Application) Prepare() error {
	a.podmanAdapter = podman.New(&podman.NewPodmanParams{
		Version:    config.PODMAN_VERSION,
		SocketPath: config.PODMAN_SOCKET,
	})

	a.systemdAdapter = systemd.New(&systemd.NewSystemdParams{
		BindPath: config.SYSTEMCTL_PATH,
		UserMode: !config.ROOTFUL,
	})

	fsq, err := quadlet.NewFilesystem(&quadlet.NewFilesystemParams{
		BaseDir:  config.QUADLET_HOME,
		UserMode: !config.ROOTFUL,
	})
	if err != nil {
		return err
	}

	a.quadletFS = fsq
	a.quadletSvc = quadlet.New(a.quadletFS, a.systemdAdapter)

	authSvc, err := auth.NewAuthService(&auth.NewAuthServiceParams{
		HostUser:   config.HOST_USER,
		Dir:        config.DATA_HOME,
		BcryptCost: 13,
	})
	if err != nil {
		return err
	}
	a.authSvc = authSvc

	return nil
}
