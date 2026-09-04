package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Port uint16 `env:"MOLESHIP_SERVER_PORT,default=5000"`

	LogLevel string `env:"MOLESHIP_LOG_LEVEL,default=info"`

	Mode string `env:"MOLESHIP_MODE"`

	ConfigHome string `env:"MOLESHIP_CONFIG_HOME"`

	CacheHome string `env:"MOLESHIP_CACHE_HOME"`

	DataHome string `env:"MOLESHIP_DATA_HOME"`

	ServerPort string `env:"MOLESHIP_SERVER_PORT,default=5000"`

	PodmanSocket string `env:"MOLESHIP_PODMAN_SOCKET"`

	SystemctlPath string `env:"MOLESHIP_BIN_SYSTEMCTL_PATH"`

	QuadletHome string `env:"MOLESHIP_QUADLET_HOME"`

	Rootful bool `env:"MOLESHIP_PODMAN_ROOTFUL_MODE,default=0"`

	PodmanVersion string `env:"MOLESHIP_PODMAN_VERSION,default=5.0.0"`

	AuthUsersStrategy string `env:"MOLESHIP_AUTH_USERS_STRATEGY,default=owner_only"`

	CORSAllowedOrigins string `env:"MOLESHIP_CORS_ALLOWED_ORIGIN,default=*"`

	JWTSecret []byte `env:"MOLESHIP_JWT_SECRET,default=supersecretkey"`
}

var defaultConfig = &Config{
	Mode:       "debug",
	ConfigHome: "",
	CacheHome:  "",
	DataHome:   "",

	ServerPort: "5000",

	PodmanSocket:  "",
	SystemctlPath: "/usr/bin/systemctl",
	QuadletHome:   "",

	Rootful: false,

	PodmanVersion: "5.0.0",

	AuthUsersStrategy: "owner_only",

	CORSAllowedOrigins: "*",

	JWTSecret: []byte("supersecretkey"),
}

func init() {
	defaultConfig = loadConfig()
}

func Current() *Config {
	return loadConfig()
}

func IsProduction() bool {
	return Current().Mode == "production" || Current().Mode == "prod"
}

func IsDebug() bool {
	return Current().Mode == "debug" || Current().Mode == "dev"
}

func loadConfig() *Config {
	cfg := &Config{}
	cfg.Mode = os.Getenv("MOLESHIP_MODE")
	cfg.LogLevel = os.Getenv("MOLESHIP_LOG_LEVEL")
	cfg.ConfigHome = os.Getenv("MOLESHIP_CONFIG_HOME")
	cfg.CacheHome = os.Getenv("MOLESHIP_CACHE_HOME")
	cfg.DataHome = os.Getenv("MOLESHIP_DATA_HOME")
	cfg.ServerPort = os.Getenv("MOLESHIP_SERVER_PORT")
	cfg.PodmanSocket = os.Getenv("MOLESHIP_PODMAN_SOCKET")
	cfg.SystemctlPath = os.Getenv("MOLESHIP_BIN_SYSTEMCTL_PATH")
	cfg.QuadletHome = os.Getenv("MOLESHIP_QUADLET_HOME")
	cfg.Rootful = os.Getenv("MOLESHIP_PODMAN_ROOTFUL_MODE") != "0"
	cfg.PodmanVersion = os.Getenv("MOLESHIP_PODMAN_VERSION")
	cfg.AuthUsersStrategy = os.Getenv("MOLESHIP_AUTH_USERS_STRATEGY")
	cfg.CORSAllowedOrigins = os.Getenv("MOLESHIP_CORS_ALLOWED_ORIGIN")

	jwtSecret := os.Getenv("MOLESHIP_JWT_SECRET")
	if jwtSecret != "" {
		cfg.JWTSecret = []byte(jwtSecret)
	}

	if cfg.Mode == "" {
		cfg.Mode = "debug"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.ConfigHome == "" {
		cfg.ConfigHome = filepath.Join(mustUserHome(), ".config", "moleship")
	}
	if cfg.CacheHome == "" {
		cfg.CacheHome = filepath.Join(mustUserHome(), ".cache", "moleship")
	}
	if cfg.DataHome == "" {
		cfg.DataHome = filepath.Join(mustUserHome(), ".local", "share", "moleship")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "5000"
	}
	if cfg.PodmanSocket == "" {
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			cfg.PodmanSocket = fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
		} else {
			cfg.PodmanSocket = filepath.Join(runtimeDir, "podman", "podman.sock")
		}
	}
	if cfg.SystemctlPath == "" {
		cfg.SystemctlPath = "/usr/bin/systemctl"
	}
	if cfg.QuadletHome == "" {
		cfg.QuadletHome = filepath.Join(mustUserHome(), ".config", "containers", "systemd")
	}
	if cfg.PodmanVersion == "" {
		cfg.PodmanVersion = "5.0.0"
	}
	if cfg.AuthUsersStrategy == "" {
		cfg.AuthUsersStrategy = "owner_only"
	}
	if cfg.CORSAllowedOrigins == "" {
		cfg.CORSAllowedOrigins = "*"
	}
	if cfg.Port == 0 {
		port, err := strconv.ParseUint(cfg.ServerPort, 10, 16)
		if err == nil {
			cfg.Port = uint16(port)
		}
	}
	if cfg.Port == 0 {
		cfg.Port = 5000
	}
	if cfg.JWTSecret == nil || len(cfg.JWTSecret) == 0 {
		cfg.JWTSecret = []byte("supersecretkey")
	}
	return cfg
}

func mustUserHome() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "/tmp"
	}
	return home
}
