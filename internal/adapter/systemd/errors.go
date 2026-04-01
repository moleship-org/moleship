package systemd

import "errors"

var (
	// ErrSystemctlNotFound indicates the systemctl binary was not found in the PATH.
	ErrSystemctlNotFound = errors.New("systemctl binary not found")

	// ErrUnitNotFound indicates the requested systemd unit does not exist.
	ErrUnitNotFound = errors.New("systemd unit not found")

	// ErrPermissionDenied indicates the operation failed due to insufficient permissions (non-rootless context).
	ErrPermissionDenied = errors.New("permission denied: ensure moleship is running in user mode")

	// ErrDaemonReloadFailed indicates the daemon-reload command failed.
	ErrDaemonReloadFailed = errors.New("failed to reload systemd daemon")

	// ErrCommandFailed indicates a generic execution error from the systemctl command.
	ErrCommandFailed = errors.New("systemd command execution failed")
)
