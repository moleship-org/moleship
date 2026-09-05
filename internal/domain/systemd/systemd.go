package systemd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const DefaultCommandTimeout = 15 * time.Second

type NewSystemdParams struct {
	BindPath       string
	UserMode       bool
	CommandTimeout time.Duration
}

type Systemd struct {
	binPath    string
	userMode   bool
	cmdTimeout time.Duration
}

var _ Port = (*Systemd)(nil)

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

	if params.CommandTimeout == 0 {
		params.CommandTimeout = DefaultCommandTimeout
	}

	sys.binPath = params.BindPath
	sys.userMode = params.UserMode
	sys.cmdTimeout = params.CommandTimeout
	return sys
}

func (s *Systemd) cmd(ctx context.Context, args ...string) (cmd *exec.Cmd, cancel context.CancelFunc) {
	ctx, cancel = context.WithTimeout(ctx, s.cmdTimeout)

	var finalArgs []string
	if s.userMode {
		finalArgs = append(finalArgs, "--user")
	}
	finalArgs = append(finalArgs, args...)

	cmd = exec.CommandContext(ctx, s.binPath, finalArgs...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	return cmd, cancel
}

func (s *Systemd) runWithStderr(ctx context.Context, args ...string) (string, error) {
	cmd, cancel := s.cmd(ctx, args...)
	defer cancel()

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stderr.String(), err
}

func (s *Systemd) UnitStatus(ctx context.Context, unitName string) (string, error) {
	cmd, cancel := s.cmd(ctx, "is-active", unitName)
	defer cancel()

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	status := strings.TrimSpace(string(out))
	stderrStr := stderr.String()

	if err != nil {
		if status == "inactive" || status == "failed" {
			return status, nil
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 4 {
			return "", ErrUnitNotFound
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
		return classifyUnitError(stderr, err)
	}
	return nil
}

func (s *Systemd) StopUnit(ctx context.Context, unitName string) error {
	stderr, err := s.runWithStderr(ctx, "stop", unitName)
	if err != nil {
		return classifyUnitError(stderr, err)
	}
	return nil
}

func (s *Systemd) RestartUnit(ctx context.Context, unitName string) error {
	stderr, err := s.runWithStderr(ctx, "restart", unitName)
	if err != nil {
		return classifyUnitError(stderr, err)
	}
	return nil
}

func (s *Systemd) ReloadDaemon(ctx context.Context) error {
	stderr, err := s.runWithStderr(ctx, "daemon-reload")
	if err != nil {
		return fmt.Errorf("%w: %v (details: %s)", ErrDaemonReloadFailed, err, stderr)
	}
	return nil
}

func classifyUnitError(stderr string, err error) error {
	if strings.Contains(stderr, "not found") || strings.Contains(stderr, "does not exist") {
		return ErrUnitNotFound
	}
	if strings.Contains(stderr, "Permission denied") || strings.Contains(stderr, "Access denied") {
		return ErrPermissionDenied
	}
	return fmt.Errorf("%w: %v (details: %s)", ErrCommandFailed, err, stderr)
}
