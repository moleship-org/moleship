package port

import (
	"context"

	"github.com/moleship-org/moleship/internal/domain/model"
)

type ContainerService interface {
	// List returns all running quadlets available on the system or an error if they cannot be retrieved.
	List(ctx context.Context, filters model.Filters) ([]model.ContainerEntity, error)

	// GetByID returns the quadlet with the given id or ErrQuadletNotFound if it does not exist.
	GetByID(ctx context.Context, id string) (*model.ContainerEntity, error)

	// GetByName returns the quadlet with the given name or ErrQuadletNotFound if it does not exist.
	GetByName(ctx context.Context, name string) (*model.ContainerEntity, error)

	// Start activates or enables the quadlet identified by name
	Start(ctx context.Context, name string) error

	// Stop deactivates or disables the quadlet identified by name.
	Stop(ctx context.Context, name string) error

	// Restart restarts the quadlet identified by name (stop then start).
	Restart(ctx context.Context, name string) error

	// Exists checks if the container is exists on podman runtime.
	Exists(ctx context.Context, name string) (bool, error)

	// Stats returns a container's resource usage.
	Stats(ctx context.Context, name string) (*model.ContainerStats, error)
}
