package systemd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type NewAdapterParams struct {
	BindPath string
	UserMode bool
}

type Adapter struct {
	binPath  string
	userMode bool
}

func New(params *NewAdapterParams) *Adapter {
	if params == nil {
		params = new(NewAdapterParams)
	}
	adr := new(Adapter)

	if params.BindPath == "" {
		path, err := exec.LookPath("systemctl")
		if err == nil {
			params.BindPath = path
		} else {
			params.BindPath = "/usr/bin/systemctl"
		}
	}

	adr.binPath = params.BindPath
	adr.userMode = params.UserMode
	return adr
}

func (a *Adapter) cmd(ctx context.Context, args ...string) *exec.Cmd {
	var finalArgs []string
	if a.userMode {
		finalArgs = append(finalArgs, "--user")
	}
	finalArgs = append(finalArgs, args...)

	return exec.CommandContext(ctx, a.binPath, finalArgs...)
}

func (a *Adapter) runWithStderr(ctx context.Context, args ...string) (string, error) {
	cmd := a.cmd(ctx, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stderr.String(), err
}

func (a *Adapter) UnitStatus(ctx context.Context, unitName string) (string, error) {
	cmd := a.cmd(ctx, "is-active", unitName)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	status := strings.TrimSpace(string(out))
	stderrStr := stderr.String()

	if err != nil {
		if status == "inactive" || status == "failed" {
			return status, nil
		}

		if strings.Contains(status, "not-found") || status == "unknown" || strings.Contains(stderrStr, "not found") {
			return "", ErrUnitNotFound
		}

		return "", fmt.Errorf("%w: %v (details: %s)", ErrCommandFailed, err, stderrStr)
	}

	return status, nil
}

func (a *Adapter) StartUnit(ctx context.Context, unitName string) error {
	stderr, err := a.runWithStderr(ctx, "start", unitName)
	if err != nil {
		if strings.Contains(stderr, "not found") || strings.Contains(stderr, "does not exist") {
			return ErrUnitNotFound
		}
		if strings.Contains(stderr, "Permission denied") {
			return ErrPermissionDenied
		}
		return fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	return nil
}

func (a *Adapter) StopUnit(ctx context.Context, unitName string) error {
	stderr, err := a.runWithStderr(ctx, "stop", unitName)
	if err != nil {
		if strings.Contains(stderr, "not found") || strings.Contains(stderr, "does not exist") {
			return ErrUnitNotFound
		}
		return fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	return nil
}

func (a *Adapter) RestartUnit(ctx context.Context, unitName string) error {
	stderr, err := a.runWithStderr(ctx, "restart", unitName)
	if err != nil {
		if strings.Contains(stderr, "not found") || strings.Contains(stderr, "does not exist") {
			return ErrUnitNotFound
		}
		return fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	return nil
}

func (a *Adapter) ReloadDaemon(ctx context.Context) error {
	stderr, err := a.runWithStderr(ctx, "daemon-reload")
	if err != nil {
		return fmt.Errorf("%w: %s (exit: %v)", ErrDaemonReloadFailed, stderr, err)
	}
	return nil
}
