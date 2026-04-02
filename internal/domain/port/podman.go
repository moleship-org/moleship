package port

import (
	"context"

	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/moleship-org/moleship/internal/domain/model"
)

type Filters map[string][]string

type PodmanSystemVersion struct {
	Data entities.ComponentVersion `json:"data,omitempty"`
}

type PodmanProvider interface {
	// Ping checks connectivity to the Podman service and returns an error if unreachable.
	Ping(ctx context.Context) error

	// GetVersion returns the podman system version component.
	GetVersion(ctx context.Context) (*PodmanSystemVersion, error)

	// ListContainers returns all the available containers with the given filters.
	ListContainers(ctx context.Context, filters Filters) ([]entities.ListContainer, error)

	// Exists determinates if a container exists.
	Exists(ctx context.Context, name string) (bool, error)

	// Stats returns a live stream of a container's resource usage.
	Stats(ctx context.Context, name string) (*model.ContainerStats, error)
}
