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
	ErrInvalidQuadlet  = errors.New("invalid quadlet definition")
	ErrQuadletNotFound = errors.New("quadlet file not found")
)

type NewQuadletServiceParams struct {
	Systemd    port.SystemdManager
	Podman     port.PodmanProvider
	QuadletDir string
}

type quadletService struct {
	systemd port.SystemdManager
	podman  port.PodmanProvider
	dir     string
}

func NewQuadletService(params *NewQuadletServiceParams) port.QuadletService {
	return &quadletService{
		systemd: params.Systemd,
		podman:  params.Podman,
		dir:     params.QuadletDir,
	}
}

func (s *quadletService) List(ctx context.Context) ([]model.Quadlet, error) {
	files, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read quadlet directory: %w", err)
	}

	realContainers, err := s.podman.ListContainers(ctx)
	if err != nil {
		realContainers = nil
	}

	quadlets := make([]model.Quadlet, 0)
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".container") {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".container")

		status, _ := s.systemd.UnitStatus(ctx, name+".service")
		q := model.Quadlet{
			Name:   name,
			Path:   filepath.Join(s.dir, f.Name()),
			Status: status,
		}

		for _, c := range realContainers {
			isMatch := false
			for _, cName := range c.Names {
				cleanName := strings.TrimPrefix(cName, "/")
				if cleanName == name || cleanName == "systemd-"+name {
					isMatch = true
					break
				}
			}

			if isMatch {
				q.Container = c
				break
			}
		}

		quadlets = append(quadlets, q)
	}

	return quadlets, nil
}

func (s *quadletService) GetByName(ctx context.Context, name string) (model.Quadlet, error) {
	all, err := s.List(ctx)
	if err != nil {
		return model.Quadlet{}, err
	}

	for _, q := range all {
		if q.Name == name {
			content, _ := os.ReadFile(q.Path)
			q.Content = string(content)
			return q, nil
		}
	}

	return model.Quadlet{}, ErrQuadletNotFound
}

func (s *quadletService) Update(ctx context.Context, name string, content string) error {
	path := filepath.Join(s.dir, name+".container")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrQuadletNotFound
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write quadlet file: %w", err)
	}

	return s.systemd.ReloadDaemon(ctx)
}

func (s *quadletService) Start(ctx context.Context, name string) error {
	return s.systemd.StartUnit(ctx, name+".service")
}

func (s *quadletService) Stop(ctx context.Context, name string) error {
	return s.systemd.StopUnit(ctx, name+".service")
}

func (s *quadletService) Restart(ctx context.Context, name string) error {
	if err := s.systemd.ReloadDaemon(ctx); err != nil {
		return err
	}
	return s.systemd.RestartUnit(ctx, name+".service")
}
