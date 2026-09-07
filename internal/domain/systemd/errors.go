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

	// ErrUnitTransientOrGenerated indicates the unit was produced by a systemd generator
	// (e.g. Podman Quadlet) and therefore cannot be enabled or disabled via systemctl.
	// Such units must have their [Install] section (WantedBy=, RequiredBy=, Alias=)
	// configured in their source definition instead.
	ErrUnitTransientOrGenerated = errors.New("unit is transient or generated and cannot be enabled/disabled directly")
)
