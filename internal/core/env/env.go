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
}

func LoadFiles(filenames ...string) error {
	return dotenv.Load(filenames...)
}

func Load() (*Env, error) {
	vars := new(Env)
	if err := env.Load(vars); err != nil {
		return nil, ErrCouldNotLoadEnvs
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrMissingHome
	}

	if vars.ConfigHome == "" {
		vars.ConfigHome = filepath.Join(home, ".config", "moleship")
	}
	if vars.CacheHome == "" {
		vars.CacheHome = filepath.Join(home, ".cache", "moleship")
	}
	if vars.DataHome == "" {
		vars.DataHome = filepath.Join(home, ".local", "share", "moleship")
	}

	dirs := []string{vars.ConfigHome, vars.CacheHome, vars.DataHome}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, ErrCouldNotCreateDirectory
		}
	}

	return vars, nil
}

func MustLoad() *Env {
	env, err := Load()
	if err != nil {
		panic(err)
	}
	return env
}
