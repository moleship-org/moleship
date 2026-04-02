package app

import (
	"io"
	"log/slog"
	"os"
	"strconv"

	"github.com/moleship-org/moleship/internal/core/env"
)

type Option func(*Config)

type Config struct {
	Mode          string
	PodmanSocket  string
	PodmanVersion string
	SystemctlPath string
	QuadletDir    string
	Port          uint16
	Rootful       bool
	Logger        *slog.Logger
}

func DefaultConfig() *Config {
	e := env.MustLoad()
	c := new(Config)

	c.PodmanSocket = e.PodmanSocket
	c.PodmanVersion = e.PodmanVersion
	c.SystemctlPath = e.SystemctlPath
	c.QuadletDir = e.QuadletHome
	c.Rootful = e.Rootful
	c.Mode = e.Mode

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
