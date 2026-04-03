package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

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

	client := &http.Client{
		Transport: transport,
		Timeout:   0,
	}

	return &Adapter{
		socketPath: params.SocketPath,
		version:    params.Version,
		libpodUri:  fmt.Sprintf("http://d/v%s/libpod", params.Version),
		client:     client,
	}
}

func (a *Adapter) getEndpoint(params ...string) string {
	uri := a.libpodUri
	for i, param := range params {
		if strings.HasPrefix(param, "?") {
			uri += strings.Join(params[i:], "")
			break
		} else {
			uri += "/" + param
		}
	}
	return uri
}

func (a *Adapter) RawCall(ctx context.Context, method string, path ...string) (*http.Response, error) {
	endpoint := a.getEndpoint(path...)
	log.Println("Endpoint", endpoint)

	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}

	if res.StatusCode >= 400 {
		defer res.Body.Close()

		var podmanErr struct {
			Cause   string `json:"cause"`
			Message string `json:"message"`
		}

		if decodeErr := json.NewDecoder(res.Body).Decode(&podmanErr); decodeErr == nil {
			return nil, fmt.Errorf("podman api error (%d): %s - %s",
				res.StatusCode, podmanErr.Cause, podmanErr.Message)
		}

		return nil, fmt.Errorf("podman api returned unexpected status: %d", res.StatusCode)
	}

	return res, nil
}
func (a *Adapter) Ping(ctx context.Context) (http.Header, error) {
	res, err := a.RawCall(ctx, http.MethodGet, "_ping")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("%w: unexpected status %s", ErrInvalidResponse, res.Status)
	}

	return res.Header, nil
}

func (a *Adapter) ListContainers(ctx context.Context, opts url.Values) ([]entities.ListContainer, error) {
	if opts == nil {
		opts = make(url.Values)
	}

	res, err := a.RawCall(ctx, http.MethodGet, "containers", "json", "?", opts.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %s", ErrInvalidResponse, res.Status)
	}

	var containers []entities.ListContainer
	if err := json.NewDecoder(res.Body).Decode(&containers); err != nil {
		return nil, ErrInvalidResponse
	}

	return containers, nil
}

func (a *Adapter) GetVersion(ctx context.Context) (*model.PodmanSystemVersion, error) {
	res, err := a.RawCall(ctx, http.MethodGet, "version")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %s", ErrInvalidResponse, res.Status)
	}

	var cv entities.ComponentVersion
	if err := json.NewDecoder(res.Body).Decode(&cv); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	pv := &model.PodmanSystemVersion{Data: cv}
	return pv, nil
}

func (a *Adapter) Exists(ctx context.Context, name string) (bool, error) {
	res, err := a.RawCall(ctx, http.MethodGet, "containers", name, "exists")
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return true, nil
	}

	return false, ErrContainerNotFound
}

func (a *Adapter) Stats(ctx context.Context, name string) (*model.ContainerStats, error) {
	res, err := a.RawCall(ctx, http.MethodGet, "containers", name, "stats", "?stream=false")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrContainerNotFound
	}
	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("podman adapter internal error")
	}

	var report model.ContainerStats
	if err := json.NewDecoder(res.Body).Decode(&report); err != nil {
		return nil, ErrInvalidResponse
	}

	return &report, nil
}

func (a *Adapter) Logs(ctx context.Context, name string, opts url.Values) (io.ReadCloser, error) {
	if opts == nil {
		opts = make(url.Values)
	}

	res, err := a.RawCall(ctx, http.MethodGet, "containers", name, "logs", "?", opts.Encode())
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrContainerNotFound
	}
	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("internal error when trying to get logs")
	}

	return res.Body, nil
}
