package podman

import "errors"

var (
	// ErrSocketNotFound indicates the podman unix socket does not exist at the provided path.
	ErrSocketNotFound = errors.New("podman socket not found")

	// ErrConnectionRefused indicates the socket exists but the podman service is not listening.
	ErrConnectionRefused = errors.New("connection refused: is podman.socket enabled?")

	// ErrContainerNotFound indicates the requested container ID or Name does not exist.
	ErrContainerNotFound = errors.New("container not found")

	// ErrAPIVersionMismatch indicates the podman api version is incompatible with the client.
	ErrAPIVersionMismatch = errors.New("podman api version mismatch")

	// ErrInvalidResponse indicates the API returned an unexpected or malformed payload.
	ErrInvalidResponse = errors.New("invalid or malformed response from podman api")
)
