package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moleship-org/moleship/internal/adapter/podman"
	"github.com/moleship-org/moleship/internal/adapter/systemd"
	"github.com/moleship-org/moleship/internal/core/api/handler"
	"github.com/moleship-org/moleship/internal/core/api/middleware"
	"github.com/moleship-org/moleship/internal/core/service"
)

type Application struct {
	cfg *Config

	router  *http.ServeMux
	handler http.Handler
}

func New(opts ...Option) *Application {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	a := new(Application)
	a.cfg = cfg
	a.router = http.NewServeMux()
	a.handler = a.router

	return a
}

func (a *Application) Start(ctx context.Context) {
	a.Prepare()

	server := &http.Server{
		Addr:    a.Addr(),
		Handler: a.handler,
	}

	serverErrors := make(chan error, 1)

	go func() {
		a.Logger().Info(fmt.Sprintf("Moleship running on http://localhost%s/ - Press CTRL+C to exit", a.Addr()))
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		a.Logger().Error(err.Error())

	case <-shutdown:
		log.Println("Starting application shutdown...")

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

func (a *Application) Logger() *slog.Logger {
	return a.cfg.Logger
}

func (a *Application) Prepare() {
	systemdAdapter := systemd.New(&systemd.NewAdapterParams{
		BindPath: os.Getenv("MOLESHIP_BIN_SYSTEMCTL_PATH"),
		UserMode: !a.cfg.Rootful,
	})

	podmanAdapter := podman.New(&podman.NewAdapterParams{
		SocketPath: os.Getenv("MOLESHIP_PODMAN_SOCKET"),
	})

	quadletSvc := service.NewQuadletService(&service.NewQuadletServiceParams{
		Systemd:    systemdAdapter,
		Podman:     podmanAdapter,
		QuadletDir: os.Getenv("MOLESHIP_QUADLET_HOME"),
	})

	handler.NewHealth().Mux(a.router)
	handler.NewQuadlet(quadletSvc).Mux(a.router)

	a.handler = middleware.Apply(a.router,
		middleware.ContextInjector,
		middleware.Logger(a.Logger()),
	)
}
