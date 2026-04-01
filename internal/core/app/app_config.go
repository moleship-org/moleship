package app

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/moleship-org/moleship/internal/core/env"
)

type Option func(*Config)

type Config struct {
	PodmanSocket  string
	SystemctlPath string
	QuadletDir    string
	Port          uint16
	Rootful       bool
	Logger        *slog.Logger
}

func DefaultConfig() *Config {
	e, err := env.Load()
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	c := new(Config)

	c.PodmanSocket = e.PodmanSocket
	c.SystemctlPath = e.SystemctlPath
	c.QuadletDir = e.QuadletHome
	c.Rootful = e.Rootful

	configPort(c, e)
	configLogger(c, e)

	return c
}

func configPort(c *Config, e *env.Env) {
	port, err := strconv.ParseUint(e.ServerPort, 10, 16)
	if err != nil {
		slog.Error("failed to parse port to uin16", "error", err)
		os.Exit(1)
	}
	if port == 0 {
		port = 6000
	}
	c.Port = uint16(port)
}

func configLogger(c *Config, e *env.Env) {
	switch e.Mode {
	case "debug-silent":
		devNull, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0)
		c.Logger = slog.New(slog.NewTextHandler(devNull, nil))
		c.Logger.Info("Using 'debug-silent' mode")

	case "production":
		debugPath := filepath.Join(e.DataHome, "journal.log")
		flog, err := os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.Error("failed to create debu file", "error", err)
			os.Exit(1)
		}

		c.Logger = slog.New(slog.NewJSONHandler(flog, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
		c.Logger.Info("Using 'production' mode")

	case "", "debug":
	default:
		c.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
		c.Logger.Info("Using 'debug' mode")
	}
}

func WithPort(port uint16) Option {
	return func(c *Config) {
		if c != nil {
			c.Port = port
		}
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(c *Config) {
		if c != nil {
			c.Logger = l
		}
	}
}

func WithPodmanSocket(socket string) Option {
	return func(c *Config) {
		if c != nil {
			c.PodmanSocket = socket
		}
	}
}

func WithQuadletDir(path string) Option {
	return func(c *Config) {
		if c != nil {
			c.QuadletDir = path
		}
	}
}

func WithSystemctlPath(path string) Option {
	return func(c *Config) {
		if c != nil {
			c.SystemctlPath = path
		}
	}
}

func WithRootful(ok bool) Option {
	return func(c *Config) {
		if c != nil {
			c.Rootful = ok
		}
	}
}
