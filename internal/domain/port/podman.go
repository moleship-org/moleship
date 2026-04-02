package port

import (
	"context"

	"github.com/containers/podman/v5/pkg/domain/entities"
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
}
