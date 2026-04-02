package service

import (
	"context"
	"errors"

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

func (s *quadletService) List(ctx context.Context) ([]model.QuadletFile, error) {
	return nil, nil
}

func (s *quadletService) Get(ctx context.Context, name string) (*model.QuadletFile, error) {
	return nil, nil
}

func (s *quadletService) Create(ctx context.Context, name string, qf *model.QuadletFile) error {
	return nil
}

func (s *quadletService) Update(ctx context.Context, override bool, name string, qf *model.QuadletFile) error {
	return nil
}

func (s *quadletService) Delete(ctx context.Context, name string) error {
	return nil
}

func (s *quadletService) Reload(ctx context.Context) error {
	return nil
}
