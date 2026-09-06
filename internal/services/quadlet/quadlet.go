package quadlet

import (
	"context"
	"errors"
	"fmt"

	"github.com/moleship-org/moleship/internal/domain/systemd"
)

type QuadletService struct {
	files   FSPort
	systemd systemd.Port
}

func New(files FSPort, sys systemd.Port) *QuadletService {
	return &QuadletService{
		files:   files,
		systemd: sys,
	}
}

type CreateOptions struct {
	FailIfExists bool

	Start bool
}

func (s *QuadletService) Create(ctx context.Context, unit Unit, opts CreateOptions) error {
	if err := s.files.Write(ctx, unit, WriteOptions{FailIfExists: opts.FailIfExists}); err != nil {
		return err
	}

	if err := s.systemd.ReloadDaemon(ctx); err != nil {
		return err
	}

	if opts.Start {
		if err := s.systemd.StartUnit(ctx, ServiceName(unit)); err != nil {
			return err
		}
	}

	return nil
}

func (s *QuadletService) Remove(ctx context.Context, kind Kind, name string) error {
	if err := validateName(name); err != nil {
		return err
	}

	serviceName := kind.ServiceName(name)

	if err := s.systemd.StopUnit(ctx, serviceName); err != nil && !errors.Is(err, systemd.ErrUnitNotFound) {
		return fmt.Errorf("failed to stop unit before removal: %w", err)
	}

	if err := s.files.Delete(ctx, kind, name); err != nil {
		return err
	}

	return s.systemd.ReloadDaemon(ctx)
}

func (s *QuadletService) Start(ctx context.Context, kind Kind, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	return s.systemd.StartUnit(ctx, kind.ServiceName(name))
}

func (s *QuadletService) Stop(ctx context.Context, kind Kind, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	return s.systemd.StopUnit(ctx, kind.ServiceName(name))
}

func (s *QuadletService) Restart(ctx context.Context, kind Kind, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	return s.systemd.RestartUnit(ctx, kind.ServiceName(name))
}

func (s *QuadletService) Status(ctx context.Context, kind Kind, name string) (string, error) {
	if err := validateName(name); err != nil {
		return "", err
	}
	return s.systemd.UnitStatus(ctx, kind.ServiceName(name))
}

func (s *QuadletService) Read(ctx context.Context, kind Kind, name string) (string, error) {
	return s.files.Read(ctx, kind, name)
}

type UnitInfo struct {
	Name        string
	Kind        Kind
	ServiceName string

	Status string

	StatusError error
}

func (s *QuadletService) List(ctx context.Context) ([]UnitInfo, error) {
	entries, err := s.files.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]UnitInfo, 0, len(entries))
	for _, e := range entries {
		serviceName := e.Kind.ServiceName(e.Name)

		info := UnitInfo{
			Name:        e.Name,
			Kind:        e.Kind,
			ServiceName: serviceName,
		}

		status, statusErr := s.systemd.UnitStatus(ctx, serviceName)
		if statusErr != nil {
			info.StatusError = statusErr
		} else {
			info.Status = status
		}

		result = append(result, info)
	}

	return result, nil
}
