package port

import (
	"context"
	"net/http"

	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/moleship-org/moleship/internal/domain/model"
)

type PodmanProvider interface {
	// RawCall makes a call to the podman socket.
	RawCall(ctx context.Context, method string, path ...string) (*http.Response, error)

	// Ping checks connectivity to the Podman service and returns an error if unreachable.
	Ping(ctx context.Context) (http.Header, error)

	// GetVersion returns the podman system version component.
	GetVersion(ctx context.Context) (*model.PodmanSystemVersion, error)

	// ListContainers returns all the available containers with the given filters.
	ListContainers(ctx context.Context, filters model.Filters) ([]entities.ListContainer, error)

	// Exists determinates if a container exists.
	Exists(ctx context.Context, name string) (bool, error)

	// Stats returns a live stream of a container's resource usage.
	Stats(ctx context.Context, name string) (*model.ContainerStats, error)
}
