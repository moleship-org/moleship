package podman

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/containers/podman/v5/pkg/domain/entities"
)

// Port provides an interface for interacting with the Podman API.
// It abstracts low-level Podman operations, including raw API calls, connectivity checks,
// version retrieval, container management, and resource monitoring.
type Port interface {
	// RawCall performs a direct HTTP call to the Podman socket API.
	RawCall(ctx context.Context, method string, path ...string) (*http.Response, error)

	// Ping checks the connectivity to the Podman service.
	Ping(ctx context.Context) (http.Header, error)

	// GetVersion returns the podman system version component.
	GetVersion(ctx context.Context) (*entities.ComponentVersion, error)

	// ListContainers returns all the available containers with the given filters.
	ListContainers(ctx context.Context, opts url.Values) ([]entities.ListContainer, error)

	// Exists determinates if a container exists.
	Exists(ctx context.Context, name string) (bool, error)

	// Stats returns a live stream of a container's resource usage.
	Stats(ctx context.Context, name string) (*entities.ContainerStatReport, error)

	// Logs returns a stream of logs.
	Logs(ctx context.Context, name string, opts url.Values) (io.ReadCloser, error)
}
