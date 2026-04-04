package port

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/moleship-org/moleship/internal/domain/model"
)

// PodmanProvider provides an interface for interacting with the Podman API.
// It abstracts low-level Podman operations, including raw API calls, connectivity checks,
// version retrieval, container management, and resource monitoring.
type PodmanProvider interface {
	// RawCall performs a direct HTTP call to the Podman socket API.
	// The method parameter specifies the HTTP method (e.g., "GET", "POST").
	// The path parameter is a variadic list of path segments to append to the base API URL.
	// Returns the HTTP response, which the caller is responsible for closing.
	// Returns an error if the request fails or if connectivity issues occur.
	RawCall(ctx context.Context, method string, path ...string) (*http.Response, error)

	// Ping checks the connectivity to the Podman service.
	// Returns the response headers from the ping request.
	// Returns an error if Podman is unreachable or not responding.
	Ping(ctx context.Context) (http.Header, error)

	// GetVersion returns the podman system version component.
	GetVersion(ctx context.Context) (*model.PodmanSystemVersion, error)

	// ListContainers returns all the available containers with the given filters.
	ListContainers(ctx context.Context, opts url.Values) ([]entities.ListContainer, error)

	// Exists determinates if a container exists.
	Exists(ctx context.Context, name string) (bool, error)

	// Stats returns a live stream of a container's resource usage.
	Stats(ctx context.Context, name string) (*model.ContainerStats, error)

	// Logs returns a stream of logs.
	Logs(ctx context.Context, name string, opts url.Values) (io.ReadCloser, error)
}
