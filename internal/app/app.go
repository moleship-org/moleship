package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

	server *http.Server
	router *chi.Mux

	podmanAdapter  podman.Port
	systemdAdapter systemd.Port
	quadletFS      quadlet.FSPort
	quadletSvc     *quadlet.QuadletService

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

func (a *Application) Addr() string {
	return fmt.Sprintf(":%d", a.cfg.Port)
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

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
		defer cancel()
		return a.server.Shutdown(shutdownCtx)
	}
}

func (a *Application) Shutdown() error {
	if a.server != nil {
		return a.server.Close()
	}
	return nil
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
	if config.IsReleaseMode() && len(config.ALLOWED_ORIGINS) == 0 {
		return errors.New("invalid cors configuration: MOLESHIP_ALLOWED_ORIGINS is required in release mode")
	}

	a.router.Use(middleware.ContextInjector(a.Logger()))
	a.router.Use(middleware.RequestLogger())

	corsOptions := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: config.IsReleaseMode(),
		MaxAge:           300,
	}
	if len(config.ALLOWED_ORIGINS) > 0 {
		corsOptions.AllowedOrigins = sanitizeAllowedOrigins(config.ALLOWED_ORIGINS)
	}
	a.router.Use(cors.Handler(corsOptions))

	a.router.Use(chi_middleware.Recoverer)
	a.router.Use(chi_middleware.RequestID)
	a.router.Use(chi_middleware.CleanPath)

	healthHandler := handler.NewHealth(a.Logger())
	authHandler := handler.NewAuth(a.Logger(), a.authSvc)
	libpodHandler := handler.NewLibpod(a.Logger(), a.podmanAdapter)
	systemdHandler := handler.NewSystemd(a.Logger(), a.systemdAdapter)
	quadletHandler := handler.NewQuadlet(a.Logger(), a.quadletSvc)

	publicLimiter := tollbooth.NewLimiter(config.PUBLIC_RATE_LIMIT, &limiter.ExpirableOptions{
		DefaultExpirationTTL: 1 * time.Hour,
	})

	publicLimiter.SetBurst(config.PUBLIC_BURST_LIMIT)
	publicLimiter.SetIPLookup(limiter.IPLookup{
		Name:           config.PUBLIC_IP_HEADER_LOOKUP,
		IndexFromRight: 0,
	})

	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(tollbooth.HTTPMiddleware(publicLimiter))

				healthHandler.Mount(r)
				authHandler.Mount(r)
			})

			r.Group(func(r chi.Router) {
				//r.Use(middleware.RequireAuth(a.authSvc))
				//r.Use(middleware.RequireCSRF())

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

func sanitizeAllowedOrigins(origins []string) []string {
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		result = append(result, origin)
	}
	return result
}
