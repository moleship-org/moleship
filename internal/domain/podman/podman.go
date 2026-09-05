package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/containers/podman/v5/pkg/domain/entities"
)

type NewPodmanParams struct {
	Version    string
	SocketPath string
}

type Podman struct {
	version    string
	socketPath string
	libpodUri  string
	client     *http.Client
}

var _ Port = (*Podman)(nil)

func New(params *NewPodmanParams) *Podman {
	if params == nil {
		params = new(NewPodmanParams)
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

	return &Podman{
		socketPath: params.SocketPath,
		version:    params.Version,
		libpodUri:  fmt.Sprintf("http://d/v%s/libpod", params.Version),
		client:     client,
	}
}

func (p *Podman) getEndpoint(params ...string) string {
	uri := p.libpodUri
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

func (p *Podman) RawCall(ctx context.Context, method string, path ...string) (*http.Response, error) {
	endpoint := p.getEndpoint(path...)

	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionRefused, err)
	}

	return res, nil
}

func decodePodmanError(res *http.Response) error {
	defer res.Body.Close()

	var podmanErr struct {
		Cause   string `json:"cause"`
		Message string `json:"message"`
	}

	if decodeErr := json.NewDecoder(res.Body).Decode(&podmanErr); decodeErr == nil {
		return fmt.Errorf("podman api error (%d): %s - %s",
			res.StatusCode, podmanErr.Cause, podmanErr.Message)
	}

	return fmt.Errorf("podman api returned unexpected status: %d", res.StatusCode)
}

func (p *Podman) Ping(ctx context.Context) (http.Header, error) {
	res, err := p.RawCall(ctx, http.MethodGet, "_ping")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, decodePodmanError(res))
	}

	return res.Header, nil
}

func (p *Podman) GetVersion(ctx context.Context) (*entities.ComponentVersion, error) {
	res, err := p.RawCall(ctx, http.MethodGet, "version")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, decodePodmanError(res))
	}

	var cv entities.ComponentVersion
	if err := json.NewDecoder(res.Body).Decode(&cv); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return &cv, nil
}

func (p *Podman) ListContainers(ctx context.Context, opts url.Values) ([]entities.ListContainer, error) {
	if opts == nil {
		opts = make(url.Values)
	}

	res, err := p.RawCall(ctx, http.MethodGet, "containers", "json", "?", opts.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, decodePodmanError(res))
	}

	var containers []entities.ListContainer
	if err := json.NewDecoder(res.Body).Decode(&containers); err != nil {
		return nil, ErrInvalidResponse
	}

	return containers, nil
}

func (p *Podman) Exists(ctx context.Context, name string) (bool, error) {
	res, err := p.RawCall(ctx, http.MethodGet, "containers", name, "exists")
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusNoContent:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, decodePodmanError(res)
	}
}

func (p *Podman) Stats(ctx context.Context, name string) (*entities.ContainerStatReport, error) {
	res, err := p.RawCall(ctx, http.MethodGet, "containers", name, "stats", "?stream=false")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, ErrContainerNotFound
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("podman adapter internal error")
	default:
		return nil, decodePodmanError(res)
	}

	var report entities.ContainerStatReport
	if err := json.NewDecoder(res.Body).Decode(&report); err != nil {
		return nil, ErrInvalidResponse
	}

	return &report, nil
}

func (p *Podman) Logs(ctx context.Context, name string, opts url.Values) (io.ReadCloser, error) {
	if opts == nil {
		opts = make(url.Values)
	}

	res, err := p.RawCall(ctx, http.MethodGet, "containers", name, "logs", "?", opts.Encode())
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		// open body to be closed by the caller
		return res.Body, nil
	case http.StatusNotFound:
		res.Body.Close()
		return nil, ErrContainerNotFound
	case http.StatusInternalServerError:
		res.Body.Close()
		return nil, fmt.Errorf("internal error when trying to get logs")
	default:
		return nil, decodePodmanError(res)
	}
}
