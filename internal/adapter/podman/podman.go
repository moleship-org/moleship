package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/moleship-org/moleship/internal/domain/port"
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

func (a *Adapter) ListContainers(ctx context.Context, filters port.Filters) ([]entities.ListContainer, error) {
	if filters == nil {
		filters = make(port.Filters)
	}

	endpoint := a.getEndpoint("containers", "json")
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w: internal url parse error", err)
	}

	q := u.Query()
	q.Set("all", "true")
	if len(filters) > 0 {
		for key, value := range filters {
			q.Set(key, strings.Join(value, ","))
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %s", ErrInvalidResponse, resp.Status)
	}

	var containers []entities.ListContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return containers, nil
}

func (a *Adapter) GetVersion(ctx context.Context) (*port.PodmanSystemVersion, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", a.getEndpoint("version"), nil)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %s", ErrInvalidResponse, resp.Status)
	}

	var v entities.ComponentVersion
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	res := &port.PodmanSystemVersion{Data: v}
	return res, nil
}
