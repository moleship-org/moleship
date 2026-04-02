package port

import (
	"context"

	"github.com/moleship-org/moleship/internal/domain/model"
)

type ContainerService interface {
	// List returns all running quadlets available on the system or an error if they cannot be retrieved.
	List(ctx context.Context, filters Filters) ([]model.QuadletEntity, error)

	// GetByID returns the quadlet with the given id or ErrQuadletNotFound if it does not exist.
	GetByID(ctx context.Context, id string) (*model.QuadletEntity, error)

	// GetByName returns the quadlet with the given name or ErrQuadletNotFound if it does not exist.
	GetByName(ctx context.Context, name string) (*model.QuadletEntity, error)

	// Start activates or enables the quadlet identified by name
	Start(ctx context.Context, name string) error

	// Stop deactivates or disables the quadlet identified by name.
	Stop(ctx context.Context, name string) error

	// Restart restarts the quadlet identified by name (stop then start).
	Restart(ctx context.Context, name string) error
}
