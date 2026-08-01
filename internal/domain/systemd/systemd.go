package systemd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type NewSystemdParams struct {
	BindPath string
	UserMode bool
}

type Systemd struct {
	binPath  string
	userMode bool
}

func New(params *NewSystemdParams) *Systemd {
	if params == nil {
		params = new(NewSystemdParams)
	}
	sys := new(Systemd)

	if params.BindPath == "" {
		path, err := exec.LookPath("systemctl")
		if err == nil {
			params.BindPath = path
		} else {
			params.BindPath = "/usr/bin/systemctl"
		}
	}

	sys.binPath = params.BindPath
	sys.userMode = params.UserMode
	return sys
}

func (s *Systemd) cmd(ctx context.Context, args ...string) *exec.Cmd {
	var finalArgs []string
	if s.userMode {
		finalArgs = append(finalArgs, "--user")
	}
	finalArgs = append(finalArgs, args...)

	return exec.CommandContext(ctx, s.binPath, finalArgs...)
}

func (s *Systemd) runWithStderr(ctx context.Context, args ...string) (string, error) {
	cmd := s.cmd(ctx, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stderr.String(), err
}

func (s *Systemd) UnitStatus(ctx context.Context, unitName string) (string, error) {
	cmd := s.cmd(ctx, "is-active", unitName)

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

func (s *Systemd) StartUnit(ctx context.Context, unitName string) error {
	stderr, err := s.runWithStderr(ctx, "start", unitName)
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

func (s *Systemd) StopUnit(ctx context.Context, unitName string) error {
	stderr, err := s.runWithStderr(ctx, "stop", unitName)
	if err != nil {
		if strings.Contains(stderr, "not found") || strings.Contains(stderr, "does not exist") {
			return ErrUnitNotFound
		}
		return fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	return nil
}

func (s *Systemd) RestartUnit(ctx context.Context, unitName string) error {
	stderr, err := s.runWithStderr(ctx, "restart", unitName)
	if err != nil {
		if strings.Contains(stderr, "not found") || strings.Contains(stderr, "does not exist") {
			return ErrUnitNotFound
		}
		return fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	return nil
}

func (s *Systemd) ReloadDaemon(ctx context.Context) error {
	stderr, err := s.runWithStderr(ctx, "daemon-reload")
	if err != nil {
		return fmt.Errorf("%w: %s (exit: %v)", ErrDaemonReloadFailed, stderr, err)
	}
	return nil
}
