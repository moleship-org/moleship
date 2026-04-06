package app

import (
	"io"
	"log/slog"
	"os"
	"strconv"

	"github.com/moleship-org/moleship/internal/core/env"
)

const DefaultPort = 5000

type Option func(*Config)

type Config struct {
	Vars   *env.Env
	Logger *slog.Logger
	Port   uint16
}

func DefaultConfig() *Config {
	e := env.MustLoad()
	c := new(Config)

	c.Vars = e
	configLogger(c, e)
	configPort(c, e)

	return c
}

func configPort(c *Config, e *env.Env) {
	port, err := strconv.ParseUint(e.ServerPort, 10, 16)
	if err != nil {
		slog.Error("failed to parse port to uin16", "error", err)
		os.Exit(1)
	}
	if port == 0 {
		port = DefaultPort
	}
	c.Port = uint16(port)
}

func configLogger(c *Config, e *env.Env) {
	var logger *slog.Logger
	switch e.Mode {
	case "silent":
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	case "production":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

	case "debug":
		fallthrough
	default:
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	c.Logger = logger
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
			c.Vars.PodmanSocket = socket
		}
	}
}

func WithQuadletHome(path string) Option {
	return func(c *Config) {
		if c != nil {
			c.Vars.QuadletHome = path
		}
	}
}

func WithSystemctlPath(path string) Option {
	return func(c *Config) {
		if c != nil {
			c.Vars.SystemctlPath = path
		}
	}
}

func WithRootful(ok bool) Option {
	return func(c *Config) {
		if c != nil {
			c.Vars.Rootful = ok
		}
	}
}
