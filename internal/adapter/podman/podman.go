package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/containers/podman/v5/pkg/domain/entities"
)

type NewAdapterParams struct {
	Version    string
	SocketPath string
}

type Adapter struct {
	version    string
	socketPath string
	libpodUri  string
	client     *http.Client
}

func New(params *NewAdapterParams) *Adapter {
	if params == nil {
		params = new(NewAdapterParams)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return net.Dial("unix", params.SocketPath)
		},
	}

	return &Adapter{
		socketPath: params.SocketPath,
		version:    params.Version,
		libpodUri:  fmt.Sprintf("http://d/v%s/libpod", params.Version),
		client:     &http.Client{Transport: transport, Timeout: 5 * time.Second},
	}
}

func (a *Adapter) getEndpoint(params ...string) string {
	uri := a.libpodUri
	for _, param := range params {
		uri = uri + "/" + param
	}
	fmt.Println(uri)
	return uri
}

func (a *Adapter) Ping(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", a.getEndpoint("_ping"), nil)
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: unexpected status %s", ErrInvalidResponse, resp.Status)
	}
	return nil
}

func (a *Adapter) ListContainers(ctx context.Context) ([]entities.ListContainer, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", a.getEndpoint("containers", "json?all=true"), nil)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("%w: server-side error", ErrInvalidResponse)
	default:
		return nil, fmt.Errorf("%w: status %s", ErrInvalidResponse, resp.Status)
	}

	var containers []entities.ListContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return containers, nil
}

func (a *Adapter) GetVersion(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", a.getEndpoint("version"), nil)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %s", ErrInvalidResponse, resp.Status)
	}

	var v podmanVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return v.Version, nil
}
