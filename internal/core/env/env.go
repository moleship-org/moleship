package env

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/ungo/env"
	"codeberg.org/ungo/env/dotenv"
)

var (
	ErrCouldNotLoadEnvs        = fmt.Errorf("could not load envs")
	ErrMissingHome             = fmt.Errorf("missing user home directory")
	ErrCouldNotCreateDirectory = fmt.Errorf("could not create directory")
)

type Env struct {
	Mode string `env:"MOLESHIP_MODE"`

	ConfigHome string `env:"MOLESHIP_CONFIG_HOME"`

	CacheHome string `env:"MOLESHIP_CACHE_HOME"`

	DataHome string `env:"MOLESHIP_DATA_HOME"`

	ServerPort string `env:"MOLESHIP_SERVER_PORT,default=6000"`

	PodmanSocket string `env:"MOLESHIP_PODMAN_SOCKET"`

	SystemctlPath string `env:"MOLESHIP_BIN_SYSTEMCTL_PATH"`

	QuadletHome string `env:"MOLESHIP_QUADLET_HOME"`

	Rootful bool `env:"MOLESHIP_ROOTFUL_MODE"`
}

func LoadFiles(filenames ...string) error {
	return dotenv.Load(filenames...)
}

func Load() (*Env, error) {
	e := new(Env)
	if err := env.Load(e); err != nil {
		return nil, ErrCouldNotLoadEnvs
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrMissingHome
	}

	if e.ConfigHome == "" {
		e.ConfigHome = filepath.Join(home, ".config", "moleship")
	}
	if e.CacheHome == "" {
		e.CacheHome = filepath.Join(home, ".cache", "moleship")
	}
	if e.DataHome == "" {
		e.DataHome = filepath.Join(home, ".local", "share", "moleship")
	}
	if e.QuadletHome == "" {
		e.QuadletHome = filepath.Join(home, ".config", "containers", "systemd")
	}

	dirs := []string{e.ConfigHome, e.CacheHome, e.DataHome, e.QuadletHome}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, ErrCouldNotCreateDirectory
		}
	}

	if e.PodmanSocket == "" {
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			e.PodmanSocket = fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
		} else {
			e.PodmanSocket = filepath.Join(runtimeDir, "podman", "podman.sock")
		}
	}

	if e.SystemctlPath == "" {
		e.SystemctlPath = "/usr/bin/systemctl"
	}

	return e, nil
}

func MustLoad() *Env {
	env, err := Load()
	if err != nil {
		panic(err)
	}
	return env
}
