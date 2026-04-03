package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/moleship-org/moleship/internal/domain/model"
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
		if strings.HasPrefix(param, "?") {
			uri += param
			break
		} else {
			uri += "/" + param
		}
	}
	return uri
}

func (a *Adapter) RawCall(ctx context.Context, method string, path ...string) (*http.Response, error) {
	endpoint := a.getEndpoint(path...)

	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := a.client.Do(req)
	if err != nil {
		return nil, ErrConnectionRefused
	}

	return res, nil
}

func (a *Adapter) Ping(ctx context.Context) (http.Header, error) {
	res, err := a.RawCall(ctx, "GET", "_ping")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("%w: unexpected status %s", ErrInvalidResponse, res.Status)
	}

	return res.Header, nil
}

func (a *Adapter) ListContainers(ctx context.Context, filters model.Filters) ([]entities.ListContainer, error) {
	if filters == nil {
		filters = make(model.Filters)
	}

	res, err := a.RawCall(ctx, "GET", "containers", "json", filters.Query())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %s", ErrInvalidResponse, res.Status)
	}

	var containers []entities.ListContainer
	if err := json.NewDecoder(res.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return containers, nil
}

func (a *Adapter) GetVersion(ctx context.Context) (*model.PodmanSystemVersion, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.getEndpoint("version"), nil)
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

	var v entities.ComponentVersion
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	res := &model.PodmanSystemVersion{Data: v}
	return res, nil
}

func (a *Adapter) Exists(ctx context.Context, name string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.getEndpoint("containers", "systemd-"+name, "exists"), nil)
	if err != nil {
		return false, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return true, nil
	}

	return false, ErrContainerNotFound
}

func (a *Adapter) Stats(ctx context.Context, name string) (*model.ContainerStats, error) {
	endpoint := a.getEndpoint("containers", "systemd-"+name, "stats") + "?stream=false"

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error de podman: código %d", resp.StatusCode)
	}

	var report model.ContainerStats
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, fmt.Errorf("error decodificando stats: %v", err)
	}

	return &report, nil
}
