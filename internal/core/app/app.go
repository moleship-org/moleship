package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/moleship-org/moleship/internal/adapter/podman"
	"github.com/moleship-org/moleship/internal/adapter/systemd"
	"github.com/moleship-org/moleship/internal/core/api/handler"
	"github.com/moleship-org/moleship/internal/core/api/middleware"
	"github.com/moleship-org/moleship/internal/core/service"
)

type Application struct {
	cfg *Config

	router chi.Router
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

func (a *Application) Start(ctx context.Context) {
	a.Prepare()

	server := &http.Server{
		Addr:         a.Addr(),
		Handler:      a.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		a.Logger().Info(fmt.Sprintf("Application running on http://localhost%s/ - Press CTRL+C to exit", a.Addr()))
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		a.Logger().Error(err.Error())

	case <-shutdown:
		a.Logger().Warn("Starting application shutdown...")

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			a.Logger().Error(err.Error())
			_ = server.Close()
		}
	}
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

func (a *Application) Prepare() {
	systemdAdapter := systemd.New(&systemd.NewAdapterParams{
		BindPath: a.cfg.SystemctlPath,
		UserMode: !a.cfg.Rootful,
	})

	podmanAdapter := podman.New(&podman.NewAdapterParams{
		SocketPath: a.cfg.PodmanSocket,
		Version:    a.cfg.PodmanVersion,
	})

	containerSvc := service.NewContainerService(&service.NewContainerServiceParams{
		Systemd:    systemdAdapter,
		Podman:     podmanAdapter,
		QuadletDir: a.cfg.QuadletDir,
	})

	quadletSvc := service.NewQuadletService(&service.NewQuadletServiceParams{
		Systemd:    systemdAdapter,
		Podman:     podmanAdapter,
		QuadletDir: a.cfg.QuadletDir,
	})

	a.router.Use(middleware.ContextInjector(a.Logger()))
	a.router.Use(middleware.Logger(a.Logger()))
	a.router.Use(chi_middleware.Recoverer)
	a.router.Use(chi_middleware.RequestID)
	a.router.Use(chi_middleware.RealIP)
	a.router.Use(chi_middleware.Timeout(60 * time.Second))

	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			handler.NewHealth().Mux(r)
			handler.NewContainer(containerSvc).Mux(r)
			handler.NewQuadlet(quadletSvc).Mux(r)
			handler.NewLibpod(podmanAdapter).Mux(r)
		})
	})
}
