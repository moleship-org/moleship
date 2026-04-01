package port

import (
	"context"

	"github.com/moleship-org/moleship/internal/domain/model"
)

type QuadletService interface {
	// List returns all quadlets available on the system or an error if they cannot be retrieved.
	List(ctx context.Context) ([]model.Quadlet, error)

	// GetByName returns the quadlet with the given name or ErrQuadletNotFound if it does not exist.
	// The returned Quadlet may include the content field depending on implementation.
	GetByName(ctx context.Context, name string) (model.Quadlet, error)

	// Start activates or enables the quadlet identified by name. Returns an error if the operation fails.
	Start(ctx context.Context, name string) error

	// Stop deactivates or disables the quadlet identified by name. Returns an error if the operation fails.
	Stop(ctx context.Context, name string) error

	// Restart restarts the quadlet identified by name (stop then start). Returns an error if the operation fails.
	Restart(ctx context.Context, name string) error
}
