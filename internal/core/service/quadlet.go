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
	"gopkg.in/ini.v1"
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
	// (~/.config/containers/systemd)
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read quadlet directory: %w", err)
	}

	var quadlets []model.QuadletFile

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".container") {
			continue
		}
		filePath := filepath.Join(s.dir, entry.Name())

		iniFile, err := ini.LoadSources(ini.LoadOptions{AllowShadows: true}, filePath)
		if err != nil {
			continue
		}

		qf := new(model.QuadletFile)
		qf.Name = strings.TrimSuffix(entry.Name(), ".container")
		qf.Path = filePath

		_ = iniFile.Section("Unit").MapTo(&qf.Unit)
		_ = iniFile.Section("Service").MapTo(&qf.Service)
		_ = iniFile.Section("Container").MapTo(&qf.Container)
		_ = iniFile.Section("Install").MapTo(&qf.Install)

		quadlets = append(quadlets, *qf)
	}

	return quadlets, nil
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
