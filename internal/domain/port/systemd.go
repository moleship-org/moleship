package port

import "context"

// SystemdManager provides an interface for interacting with systemd units.
// It abstracts the management of systemd services, allowing operations such as
// checking status, starting, stopping, restarting units, and reloading the daemon.
type SystemdManager interface {
	// UnitStatus retrieves the current status of the specified systemd unit.
	// The unitName parameter should be the full unit name (e.g., "nginx.service", "myapp.timer").
	// It returns the status as a string, such as "active", "inactive", "failed", "loaded", etc.
	// Returns an error if the unit does not exist, if permission is denied, or if the query fails.
	UnitStatus(ctx context.Context, unitName string) (string, error)

	// StartUnit starts the specified systemd unit. Returns an error if the start operation fails.
	StartUnit(ctx context.Context, unitName string) error

	// StopUnit stops the specified systemd unit. Returns an error if the stop operation fails.
	StopUnit(ctx context.Context, unitName string) error

	// RestartUnit restarts the specified systemd unit. Returns an error if the restart operation fails.
	RestartUnit(ctx context.Context, unitName string) error

	// ReloadDaemon reloads the systemd manager configuration (equivalent to `systemctl daemon-reload`).
	// Returns an error if the reload operation fails.
	ReloadDaemon(ctx context.Context) error
}
