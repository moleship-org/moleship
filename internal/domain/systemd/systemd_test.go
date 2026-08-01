package systemd_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/moleship-org/moleship/internal/adapter/systemd"
)

func TestMain(m *testing.M) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		handleMockSystemctl()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func handleMockSystemctl() {
	args := os.Args

	isUser := false
	command := ""
	unit := ""

	for i, arg := range args {
		if arg == "--user" {
			isUser = true
		}
		if arg == "is-active" || arg == "start" || arg == "stop" {
			command = arg
			if i+1 < len(args) {
				unit = args[i+1]
			}
		}
	}

	if !isUser {
		fmt.Fprint(os.Stderr, "Error: missing --user flag")
		os.Exit(1)
	}

	switch command {
	case "is-active":
		if unit == "valid.service" {
			fmt.Print("active")
			os.Exit(0)
		}
		fmt.Print("unknown")
		os.Exit(4) // Código de systemd para unit not found
	case "start":
		if unit == "invalid.service" {
			fmt.Fprint(os.Stderr, "unit not found")
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(0)
}

func TestAdapter_UnitStatus(t *testing.T) {
	bin, _ := os.Executable()
	adapter := systemd.New(&systemd.NewAdapterParams{
		BindPath: bin,
		UserMode: true,
	})

	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	defer os.Unsetenv("GO_WANT_HELPER_PROCESS")

	ctx := context.Background()

	t.Run("Active Unit", func(t *testing.T) {
		status, err := adapter.UnitStatus(ctx, "valid.service")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if status != "active" {
			t.Errorf("expected active, got %s", status)
		}
	})

	t.Run("Unit Not Found", func(t *testing.T) {
		_, err := adapter.UnitStatus(ctx, "nonexistent.service")
		if !errors.Is(err, systemd.ErrUnitNotFound) {
			t.Errorf("expected ErrUnitNotFound, got %v", err)
		}
	})
}

func TestAdapter_StartUnit(t *testing.T) {
	bin, _ := os.Executable()
	adapter := systemd.New(&systemd.NewAdapterParams{
		BindPath: bin,
		UserMode: true,
	})

	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	defer os.Unsetenv("GO_WANT_HELPER_PROCESS")

	ctx := context.Background()

	t.Run("Success Start", func(t *testing.T) {
		err := adapter.StartUnit(ctx, "valid.service")
		if err != nil {
			t.Fatalf("should not fail: %v", err)
		}
	})

	t.Run("Fail Start Unit Not Found", func(t *testing.T) {
		err := adapter.StartUnit(ctx, "invalid.service")
		if !errors.Is(err, systemd.ErrUnitNotFound) {
			t.Errorf("expected ErrUnitNotFound, got %v", err)
		}
	})
}
