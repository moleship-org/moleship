package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/moleship-org/moleship/internal/config"
)

const (
	DefaultReadTimeout       = 20 * time.Second
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultWriteTimeout      = 30 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultMaxHeaderBytes    = 1024 * 1024 // 1 MB
	DefaultShutdownTimeout   = 5 * time.Second
)

type Option func(*Config)

type Config struct {
	Port uint16

	Logger *slog.Logger

	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int

	ShutdownTimeout time.Duration
}

func DefaultConfig() *Config {
	cfg := new(Config)
	cfg.Port = 5000

	var logHandler slog.Handler
	switch config.LOG_LEVEL {
	case "debug":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "warn":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn})
	case "error":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})
	case "info":
		fallthrough
	default:
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	cfg.Logger = slog.New(logHandler)

	cfg.ReadTimeout = DefaultReadTimeout
	cfg.ReadHeaderTimeout = DefaultReadHeaderTimeout
	cfg.WriteTimeout = DefaultWriteTimeout
	cfg.IdleTimeout = DefaultIdleTimeout
	cfg.MaxHeaderBytes = DefaultMaxHeaderBytes
	cfg.ShutdownTimeout = DefaultShutdownTimeout

	return cfg
}

func WithPort(port uint16) Option {
	return func(c *Config) {
		c.Port = port
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

func WithReadHeaderTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadHeaderTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.IdleTimeout = timeout
	}
}

func WithMaxHeaderBytes(maxBytes int) Option {
	return func(c *Config) {
		c.MaxHeaderBytes = maxBytes
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ShutdownTimeout = timeout
	}
}
