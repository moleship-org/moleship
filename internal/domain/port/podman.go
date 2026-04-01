package port

import (
	"context"

	"github.com/containers/podman/v5/pkg/domain/entities"
)

type PodmanProvider interface {
	// Ping checks connectivity to the Podman service and returns an error if unreachable.
	Ping(ctx context.Context) error

	// ListContainers returns a slice of ContainerInfo representing all containers known to Podman,
	// or an error if the list cannot be retrieved.
	ListContainers(ctx context.Context) ([]entities.ListContainer, error)

	// GetVersion returns the Podman API/version string (e.g., "4.5.0") or an error if it cannot be determined.
	GetVersion(ctx context.Context) (string, error)
}
