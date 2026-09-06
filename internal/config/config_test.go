package config

import (
	"path/filepath"
	"testing"
)

func TestHandleRootlessConfigUsesHostUserAndAbsoluteHomePath(t *testing.T) {
	oldHostUser := HOST_USER
	oldConfigHome := CONFIG_HOME
	oldCacheHome := CACHE_HOME
	oldDataHome := DATA_HOME
	oldQuadletHome := QUADLET_HOME
	defer func() {
		HOST_USER = oldHostUser
		CONFIG_HOME = oldConfigHome
		CACHE_HOME = oldCacheHome
		DATA_HOME = oldDataHome
		QUADLET_HOME = oldQuadletHome
	}()

	HOST_USER = "nicolito"
	t.Setenv(EnvConfigHome, "")
	t.Setenv(EnvCacheHome, "")
	t.Setenv(EnvDataHome, "")
	t.Setenv(EnvQuadletHome, "")

	handleRootlessConfig()

	homeDir := filepath.Join(string(filepath.Separator), "home", "nicolito")
	if got, want := CONFIG_HOME, filepath.Join(homeDir, ".config", "moleship"); got != want {
		t.Fatalf("CONFIG_HOME = %q, want %q", got, want)
	}
	if got, want := CACHE_HOME, filepath.Join(homeDir, ".cache", "moleship"); got != want {
		t.Fatalf("CACHE_HOME = %q, want %q", got, want)
	}
	if got, want := DATA_HOME, filepath.Join(homeDir, ".local", "share", "moleship"); got != want {
		t.Fatalf("DATA_HOME = %q, want %q", got, want)
	}
	if got, want := QUADLET_HOME, filepath.Join(homeDir, ".config", "containers", "systemd"); got != want {
		t.Fatalf("QUADLET_HOME = %q, want %q", got, want)
	}
}

func TestHandleRootlessConfigHonorsExplicitOverrides(t *testing.T) {
	oldHostUser := HOST_USER
	oldConfigHome := CONFIG_HOME
	oldCacheHome := CACHE_HOME
	oldDataHome := DATA_HOME
	oldQuadletHome := QUADLET_HOME
	defer func() {
		HOST_USER = oldHostUser
		CONFIG_HOME = oldConfigHome
		CACHE_HOME = oldCacheHome
		DATA_HOME = oldDataHome
		QUADLET_HOME = oldQuadletHome
	}()

	HOST_USER = "ignored"
	t.Setenv(EnvConfigHome, "/tmp/custom-config")
	t.Setenv(EnvCacheHome, "/tmp/custom-cache")
	t.Setenv(EnvDataHome, "/tmp/custom-data")
	t.Setenv(EnvQuadletHome, "/tmp/custom-quadlet")

	handleRootlessConfig()

	if got, want := CONFIG_HOME, "/tmp/custom-config"; got != want {
		t.Fatalf("CONFIG_HOME = %q, want %q", got, want)
	}
	if got, want := CACHE_HOME, "/tmp/custom-cache"; got != want {
		t.Fatalf("CACHE_HOME = %q, want %q", got, want)
	}
	if got, want := DATA_HOME, "/tmp/custom-data"; got != want {
		t.Fatalf("DATA_HOME = %q, want %q", got, want)
	}
	if got, want := QUADLET_HOME, "/tmp/custom-quadlet"; got != want {
		t.Fatalf("QUADLET_HOME = %q, want %q", got, want)
	}
}
