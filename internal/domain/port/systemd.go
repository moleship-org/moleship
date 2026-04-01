package port

import "context"

type SystemdManager interface {
	// UnitStatus returns the current status of the specified systemd unit as a string
	// (e.g., "active", "inactive", "failed") or an error if the status cannot be determined.
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
