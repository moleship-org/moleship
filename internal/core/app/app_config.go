package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/moleship-org/moleship/internal/core/env"
)

type Option func(*Config)

type Config struct {
	PodmanSocket  string
	PodmanPath    string
	SystemctlPath string
	Port          uint16
	Rootful       bool
	Logger        *slog.Logger
}

func DefaultConfig() *Config {
	c := new(Config)

	vars, err := env.Load()
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	port, err := strconv.ParseUint(vars.ServerPort, 10, 16)
	if err != nil {
		slog.Error("failed to parse port to uin16", "error", err)
		os.Exit(1)
	}
	if port == 0 {
		port = 6000
	}
	c.Port = uint16(port)

	socket := vars.PodmanSocket
	if socket == "" {
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			socket = fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
		} else {
			socket = filepath.Join(runtimeDir, "podman", "podman.sock")
		}
	}
	c.PodmanSocket = socket

	mode := os.Getenv("MOLESHIP_MODE")
	switch mode {
	case "debug-silent":
		devNull, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0)
		c.Logger = slog.New(slog.NewTextHandler(devNull, nil))
		c.Logger.Info("Using 'debug-silent' mode")

	case "production":
		flog, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.Error("failed to create debu file", "error", err)
			os.Exit(1)
		}

		c.Logger = slog.New(slog.NewJSONHandler(flog, nil))
		c.Logger.Info("Using 'production' mode")

	case "", "debug":
	default:
		c.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
		c.Logger.Info("Using 'debug' mode")
	}

	c.SystemctlPath = os.Getenv("MOLESHIP_BIN_SYSTEMCTL_PATH")
	c.PodmanPath = os.Getenv("MOLESHIP_BIN_PODMAN_PATH")

	return c
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

func WithPodmanPath(path string) Option {
	return func(c *Config) {
		if c != nil {
			c.PodmanPath = path
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
