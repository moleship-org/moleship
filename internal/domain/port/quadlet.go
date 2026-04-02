package port

import (
	"context"

	"github.com/moleship-org/moleship/internal/domain/model"
)

type QuadletService interface {
	// List returns all quadlets files on the system.
	List(ctx context.Context) ([]model.QuadletFile, error)

	// Get returns a quadlet unit file information if it exists.
	Get(ctx context.Context, name string) (*model.QuadletFile, error)

	// Create creates a new quadlet file with the given name and options.
	Create(ctx context.Context, name string, qf *model.QuadletFile) error

	// Update updates a quadlet file options.
	Update(ctx context.Context, override bool, name string, qf *model.QuadletFile) error

	// Delete removes a quadlet file from the system.
	Delete(ctx context.Context, name string) error

	// Reload reloads the systemd daemons.
	Reload(ctx context.Context) error
}
