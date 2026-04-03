package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/moleship-org/moleship/internal/adapter/podman"
	"github.com/moleship-org/moleship/internal/domain/model"
	"github.com/moleship-org/moleship/internal/domain/port"
)

var (
	ErrInvalidContainer   = errors.New("invalid container definition")
	ErrContainertNotFound = errors.New("container not found")
)

type NewContainerServiceParams struct {
	Systemd    port.SystemdManager
	Podman     port.PodmanProvider
	QuadletDir string
}

type containerServiceImpl struct {
	systemd port.SystemdManager
	podman  port.PodmanProvider
	dir     string
}

func NewContainerService(params *NewContainerServiceParams) port.ContainerService {
	return &containerServiceImpl{
		systemd: params.Systemd,
		podman:  params.Podman,
		dir:     params.QuadletDir,
	}
}

func (s *containerServiceImpl) List(ctx context.Context, opts url.Values) ([]model.ContainerEntity, error) {
	if opts == nil {
		opts = make(url.Values)
	}

	files, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read quadlet directory: %w", err)
	}

	quadlets := make([]model.ContainerEntity, 0)
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".container") {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".container")

		status, _ := s.systemd.UnitStatus(ctx, name+".service")
		q := model.ContainerEntity{
			Name:   name,
			Path:   filepath.Join(s.dir, f.Name()),
			Status: status,
		}

		opts["name"] = []string{s.sanitizeName(name)}
		containers, err := s.podman.ListContainers(ctx, opts)
		if err == nil {
			if len(containers) >= 1 {
				q.Container = containers[0]
			}
		}

		quadlets = append(quadlets, q)
	}

	return quadlets, nil
}

func (s *containerServiceImpl) GetByID(ctx context.Context, id string) (*model.ContainerEntity, error) {
	opts := url.Values{
		"id": {id},
	}

	containers, err := s.List(ctx, opts)
	if err != nil {
		return nil, ErrInvalidContainer
	}

	var q *model.ContainerEntity
	if len(containers) >= 1 {
		q = &containers[0]
	}

	return q, nil
}

func (s *containerServiceImpl) GetByName(ctx context.Context, name string) (*model.ContainerEntity, error) {
	fileName := s.getPlainName(name) + ".container"
	path := filepath.Join(s.dir, fileName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrContainertNotFound
	}

	status, _ := s.systemd.UnitStatus(ctx, name+".service")
	q := &model.ContainerEntity{
		Name:   s.getPlainName(name),
		Path:   path,
		Status: status,
	}

	filters := map[string][]string{
		"name": {s.getPlainName(name), s.sanitizeName(name)},
	}
	b, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("marshal filters")
	}

	opts := url.Values{
		"filters": {string(b)},
	}

	containers, err := s.podman.ListContainers(ctx, opts)
	if errors.Is(err, podman.ErrContainerNotFound) {
		return nil, ErrContainertNotFound
	}
	if len(containers) >= 1 {
		q.Container = containers[0]
	}

	return q, nil
}

func (s *containerServiceImpl) Start(ctx context.Context, name string) error {
	return s.systemd.StartUnit(ctx, name+".service")
}

func (s *containerServiceImpl) Stop(ctx context.Context, name string) error {
	return s.systemd.StopUnit(ctx, name+".service")
}

func (s *containerServiceImpl) Restart(ctx context.Context, name string) error {
	if err := s.systemd.ReloadDaemon(ctx); err != nil {
		return err
	}
	return s.systemd.RestartUnit(ctx, name+".service")
}

func (s *containerServiceImpl) Exists(ctx context.Context, name string) (bool, error) {
	ok, err := s.podman.Exists(ctx, s.sanitizeName(name))
	if errors.Is(err, podman.ErrContainerNotFound) {
		return false, ErrContainertNotFound
	}
	return ok, err
}

func (s *containerServiceImpl) Stats(ctx context.Context, name string) (*model.ContainerStats, error) {
	report, err := s.podman.Stats(ctx, s.sanitizeName(name))
	if errors.Is(err, podman.ErrContainerNotFound) {
		return nil, ErrContainertNotFound
	}
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s *containerServiceImpl) Logs(ctx context.Context, name string, opts url.Values) (io.ReadCloser, error) {
	logs, err := s.podman.Logs(ctx, s.sanitizeName(name), opts)
	if errors.Is(err, podman.ErrContainerNotFound) {
		return nil, ErrContainertNotFound
	}
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (svc *containerServiceImpl) sanitizeName(name string) (s string) {
	s = strings.TrimSpace(name)
	if !strings.HasPrefix(s, "systemd-") {
		s = "systemd-" + s
	}
	return s
}

func (svc *containerServiceImpl) getPlainName(name string) string {
	if strings.HasPrefix(name, "systemd-") {
		s, _ := strings.CutPrefix(name, "systemd-")
		return s
	}
	return name
}
