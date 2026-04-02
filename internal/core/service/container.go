package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (s *containerServiceImpl) List(ctx context.Context, filters port.Filters) ([]model.ContainerEntity, error) {
	if filters == nil {
		filters = make(port.Filters)
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

		filters["name"] = []string{name}
		containers, err := s.podman.ListContainers(ctx, filters)
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
	filters := port.Filters{
		"id": {id},
	}

	containers, err := s.List(ctx, filters)
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
	fileName := name + ".container"
	path := filepath.Join(s.dir, fileName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrContainertNotFound
	}

	status, _ := s.systemd.UnitStatus(ctx, name+".service")
	q := &model.ContainerEntity{
		Name:   name,
		Path:   path,
		Status: status,
	}

	filters := port.Filters{
		"name": {name},
	}

	containers, err := s.podman.ListContainers(ctx, filters)
	if err == nil && len(containers) > 0 {
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
	ok, err := s.podman.Exists(ctx, name)
	if errors.Is(err, ErrContainertNotFound) {
		return false, nil
	}
	return ok, err
}

func (s *containerServiceImpl) Stats(ctx context.Context, name string) (*model.ContainerStats, error) {
	report, err := s.podman.Stats(ctx, name)
	if err != nil {
		return nil, err
	}
	return report, nil
}
