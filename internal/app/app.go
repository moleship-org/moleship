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

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/moleship-org/moleship/internal/api/handler"
	"github.com/moleship-org/moleship/internal/api/middleware"
	"github.com/moleship-org/moleship/internal/domain/config"
	"github.com/moleship-org/moleship/internal/domain/crypto"
	"github.com/moleship-org/moleship/internal/domain/persistence"
	"github.com/moleship-org/moleship/internal/domain/podman"
	"github.com/moleship-org/moleship/internal/domain/systemd"
	"github.com/moleship-org/moleship/internal/service/auth"
	"github.com/moleship-org/moleship/internal/service/container"
	"github.com/moleship-org/moleship/internal/service/quadlet"

	_ "modernc.org/sqlite"
)

type Application struct {
	cfg    *Config
	router chi.Router

	// --- Domain Services

	systemdSvc  systemd.SystemdPort
	podmanSvc   podman.PodmanPort
	passwordMan crypto.PasswordManager
	tokenGen    crypto.TokenGenerator

	// --- Services

	containerSvc *container.ContainerService
	quadletSvc   *quadlet.QuadletService
	authSvc      *auth.AuthService

	// --- Persistence

	repo        persistence.Repository
	userRepo    *persistence.UserRepository
	sessionRepo *persistence.SessionRepository
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
		ReadTimeout: a.cfg.ReadTimeout,
		// Time for reading the request headers only (helps mitigate slowloris attacks)
		ReadHeaderTimeout: a.cfg.ReadHeaderTimeout,
		// Time for writing the response
		WriteTimeout: a.cfg.WriteTimeout,
		// Time that an idle connection waits before closing
		IdleTimeout: a.cfg.IdleTimeout,
		// Maximum size of request headers (helps mitigate DoS attacks)
		MaxHeaderBytes: a.cfg.MaxHeaderBytes,
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

		ctx, cancel := context.WithTimeout(ctx, a.cfg.ShutdownTimeout)
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
	a.setupRoutes()
}

func (a *Application) setupRoutes() error {
	if a.repo == nil || a.userRepo == nil || a.sessionRepo == nil {
		return fmt.Errorf("repositories are not initialized")
	}
	if a.systemdSvc == nil || a.podmanSvc == nil || a.containerSvc == nil || a.quadletSvc == nil || a.authSvc == nil {
		return fmt.Errorf("services are not initialized")
	}

	a.router.Use(middleware.ContextInjector(a.Logger()))
	a.router.Use(middleware.Logger(a.Logger()))
	a.router.Use(middleware.CORS())
	a.router.Use(chi_middleware.Recoverer)
	a.router.Use(chi_middleware.RequestID)
	a.router.Use(chi_middleware.RealIP)

	healthHandler := handler.NewHealth()
	authHandler := handler.NewAuth(a.authSvc)
	containerHandler := handler.NewContainer(a.containerSvc)
	quadletHandler := handler.NewQuadlet(a.quadletSvc)
	libpodHandler := handler.NewLibpod(a.podmanSvc)
	userHandler := handler.NewUser(a.userRepo)
	adminHandler := handler.NewAdmin(a.userRepo)

	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			healthHandler.Mux(r)
			authHandler.Mux(r)

			// Protected routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(a.authSvc))

				containerHandler.Mux(r)
				quadletHandler.Mux(r)
				libpodHandler.Mux(r)
				userHandler.Mux(r)

				// Admin-only routes
				r.Group(func(r chi.Router) {
					r.Use(middleware.AdminOnly(a.userRepo))

					adminHandler.Mux(r)
				})
			})
		})
	})

	return nil
}

func (a *Application) setupServices() {
	a.systemdSvc = systemd.New(&systemd.NewSystemdParams{
		BindPath: config.Current().SystemctlPath,
		UserMode: !config.Current().Rootful,
	})

	a.podmanSvc = podman.New(&podman.NewPodmanParams{
		SocketPath: config.Current().PodmanSocket,
		Version:    config.Current().PodmanVersion,
	})

	a.containerSvc = container.NewContainerService(&container.NewContainerServiceParams{
		Systemd: a.systemdSvc,
		Podman:  a.podmanSvc,
	})

	a.quadletSvc = quadlet.NewQuadletService(&quadlet.NewQuadletServiceParams{
		Systemd: a.systemdSvc,
		Podman:  a.podmanSvc,
	})

	a.passwordMan = crypto.NewDefaultPasswordManager()
	a.tokenGen = crypto.NewTokenGenerator()

	a.authSvc = auth.NewAuthService(&auth.AuthServiceParams{
		UsersStrategyFlag: config.Current().AuthUsersStrategy,
		UserRepo:          a.userRepo,
		SessionRepo:       a.sessionRepo,
		PasswordManager:   a.passwordMan,
		TokenGenerator:    a.tokenGen,
	})
}

func (a *Application) setupDatabase() {
	path := fmt.Sprintf("%s/moleship.db", config.Current().DataHome)
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

	err = persistence.RunMigrations(conn, "database/migrations")
	if err != nil {
		a.Logger().Error(fmt.Sprintf("Failed to run database migrations: %s", err.Error()))
		os.Exit(1)
	}

	a.repo = persistence.NewSQLiteRepository(conn)
	a.userRepo = persistence.NewUserRepository(a.repo)
	a.sessionRepo = persistence.NewSessionRepository(a.repo)
}
