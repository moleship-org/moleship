package systemd

import "context"

// Port provides an interface for interacting with systemd units.
type Port interface {
	// UnitStatus retrieves the current status of the specified systemd unit.
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
