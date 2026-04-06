package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/moleship-org/moleship/docs"
	"github.com/moleship-org/moleship/internal/adapter/crypto"
	"github.com/moleship-org/moleship/internal/adapter/persistence"
	"github.com/moleship-org/moleship/internal/adapter/podman"
	"github.com/moleship-org/moleship/internal/adapter/systemd"
	"github.com/moleship-org/moleship/internal/core/api/handler"
	"github.com/moleship-org/moleship/internal/core/api/middleware"
	"github.com/moleship-org/moleship/internal/core/service"
	"github.com/moleship-org/moleship/internal/domain/port"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "modernc.org/sqlite"
)

const (
	DefaultReadTimeout       = 20 * time.Second
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultWriteTimeout      = 30 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
)

type Application struct {
	cfg    *Config
	router chi.Router

	// --- Services

	systemdSvc   port.SystemdManager
	podmanSvc    port.PodmanProvider
	containerSvc port.ContainerService
	quadletSvc   port.QuadletService
	authSvc      port.AuthService
	passwordMan  port.PasswordManager
	tokenGen     port.TokenGenerator

	// --- Persistence

	repo        persistence.Repository
	userRepo    port.UserRepository
	sessionRepo port.SessionRepository
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
		Addr:    a.Addr(),
		Handler: a.router,
		// Time for the whole request to complete (headers + body)
		ReadTimeout: DefaultReadTimeout,
		// Time for reading the request headers only (helps mitigate slowloris attacks)
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		// Time for writing the response
		WriteTimeout: DefaultWriteTimeout,
		// Time that an idle connection waits before closing
		IdleTimeout: DefaultIdleTimeout,
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
	a.setupDatabase()
	a.setupServices()
	a.setupRouter()
}

func (a *Application) setupDatabase() {
	path := fmt.Sprintf("%s/moleship.db", a.cfg.DataHome)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		a.Logger().Info("Database file not found, creating new one...")
		file, err := os.Create(path)
		if err != nil {
			a.Logger().Error(fmt.Sprintf("Failed to create database file: %s", err.Error()))
			os.Exit(1)
		}
		file.Close()
	}

	dbUri := fmt.Sprintf("file:%s?cache=shared&_fk=1", path)
	conn, err := sql.Open("sqlite", dbUri)
	if err != nil {
		a.Logger().Error(fmt.Sprintf("Failed to open database: %s", err.Error()))
		os.Exit(1)
	}

	err = persistence.RunMigrations(conn, "db/migrations")
	if err != nil {
		a.Logger().Error(fmt.Sprintf("Failed to run database migrations: %s", err.Error()))
		os.Exit(1)
	}

	a.repo = persistence.NewSQLiteRepository(conn)
	a.userRepo = persistence.NewUserRepository(a.repo)
	a.sessionRepo = persistence.NewSessionRepository(a.repo)
}

func (a *Application) setupServices() {
	a.systemdSvc = systemd.New(&systemd.NewAdapterParams{
		BindPath: a.cfg.SystemctlPath,
		UserMode: !a.cfg.Rootful,
	})

	a.podmanSvc = podman.New(&podman.NewAdapterParams{
		SocketPath: a.cfg.PodmanSocket,
		Version:    a.cfg.PodmanVersion,
	})

	a.containerSvc = service.NewContainerService(&service.NewContainerServiceParams{
		Systemd:    a.systemdSvc,
		Podman:     a.podmanSvc,
		QuadletDir: a.cfg.QuadletHome,
	})

	a.quadletSvc = service.NewQuadletService(&service.NewQuadletServiceParams{
		Systemd:    a.systemdSvc,
		Podman:     a.podmanSvc,
		QuadletDir: a.cfg.QuadletHome,
	})

	a.passwordMan = crypto.NewDefaultPasswordManager()
	a.tokenGen = crypto.NewTokenGenerator()

	a.authSvc = service.NewAuthService(&service.AuthServiceParams{
		UserRepo:        a.userRepo,
		SessionRepo:     a.sessionRepo,
		PasswordManager: a.passwordMan,
		TokenGenerator:  a.tokenGen,
	})
}

func (a *Application) setupRouter() {
	a.router.Use(middleware.ContextInjector(a.Logger()))
	a.router.Use(middleware.Logger(a.Logger()))
	a.router.Use(chi_middleware.Recoverer)
	a.router.Use(chi_middleware.RequestID)
	a.router.Use(chi_middleware.RealIP)
	a.router.Use(chi_middleware.Timeout(60 * time.Second))

	if a.cfg.Mode != "production" {
		a.router.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(docs.SwaggerInfo.SwaggerTemplate))
		})

		url := fmt.Sprintf("http://localhost:%d/swagger/doc.json", a.cfg.Port)
		a.router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(url)))
	}

	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			handler.NewHealth().Mux(r)
			handler.NewAuth(a.authSvc).Mux(r)

			// Protected routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(a.authSvc))
				handler.NewContainer(a.containerSvc).Mux(r)
				handler.NewQuadlet(a.quadletSvc).Mux(r)
				handler.NewLibpod(a.podmanSvc).Mux(r)
			})
		})
	})
}
