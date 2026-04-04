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

type QuadletService struct {
	systemd port.SystemdManager
	podman  port.PodmanProvider
	dir     string
}

func NewQuadletService(params *NewQuadletServiceParams) *QuadletService {
	return &QuadletService{
		systemd: params.Systemd,
		podman:  params.Podman,
		dir:     params.QuadletDir,
	}
}

func (s *QuadletService) List(ctx context.Context) ([]model.QuadletFile, error) {
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

func (s *QuadletService) Get(ctx context.Context, name string) (*model.QuadletFile, error) {
	filePath := filepath.Join(s.dir, name+".container")

	iniFile, err := ini.LoadSources(ini.LoadOptions{AllowShadows: true}, filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrQuadletNotFound
		}
		return nil, fmt.Errorf("failed to read quadlet file: %w", err)
	}

	qf := new(model.QuadletFile)
	qf.Name = name
	qf.Path = filePath

	_ = iniFile.Section("Unit").MapTo(&qf.Unit)
	_ = iniFile.Section("Service").MapTo(&qf.Service)
	_ = iniFile.Section("Container").MapTo(&qf.Container)
	_ = iniFile.Section("Install").MapTo(&qf.Install)

	return qf, nil
}

func (s *QuadletService) Exists(ctx context.Context, name string) (bool, error) {
	filePath := filepath.Join(s.dir, name+".container")

	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if quadlet file exists: %w", err)
}

func (s *QuadletService) Create(ctx context.Context, name string, qf *model.QuadletFile) error {
	filePath := filepath.Join(s.dir, name+".container")

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("quadlet file already exists: %s", name)
	}

	// Ensure directory exists
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("failed to create quadlet directory: %w", err)
	}

	// Write the quadlet file
	if err := os.WriteFile(filePath, []byte(qf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write quadlet file: %w", err)
	}

	return s.Reload(ctx)
}

func (s *QuadletService) Update(ctx context.Context, override bool, name string, qf *model.QuadletFile) error {
	filePath := filepath.Join(s.dir, name+".container")

	// Check if file exists
	existing, err := s.Get(ctx, name)
	if err != nil {
		if err == ErrQuadletNotFound && override {
			// If override=true and file doesn't exist, treat as Create
			return s.Create(ctx, name, qf)
		}
		return err
	}

	// Merge strategy: if override=false, merge configurations; if override=true, replace entirely
	var finalQf *model.QuadletFile
	if override {
		finalQf = qf
	} else {
		finalQf = mergeQuadletFiles(existing, qf)
	}

	// Write the updated quadlet file
	if err := os.WriteFile(filePath, []byte(finalQf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write quadlet file: %w", err)
	}

	return s.Reload(ctx)
}

// mergeQuadletFiles performs a smart merge of two quadlet files.
// New values override existing ones for scalars; slices are merged (union).
func mergeQuadletFiles(existing *model.QuadletFile, new *model.QuadletFile) *model.QuadletFile {
	merged := &model.QuadletFile{
		Name: existing.Name,
		Path: existing.Path,
	}

	// Merge Unit section
	merged.Unit = mergeUnitOptions(existing.Unit, new.Unit)

	// Merge Service section
	merged.Service = mergeServiceOptions(existing.Service, new.Service)

	// Merge Container section
	merged.Container = mergeContainerOptions(existing.Container, new.Container)

	// Merge Install section
	merged.Install = mergeInstallOptions(existing.Install, new.Install)

	return merged
}

func mergeUnitOptions(existing, new model.UnitOptions) model.UnitOptions {
	merged := existing
	if new.Description != "" {
		merged.Description = new.Description
	}
	merged.Requires = mergeSlices(merged.Requires, new.Requires)
	merged.Wants = mergeSlices(merged.Wants, new.Wants)
	merged.After = mergeSlices(merged.After, new.After)
	merged.Before = mergeSlices(merged.Before, new.Before)
	return merged
}

func mergeServiceOptions(existing, new model.ServiceOptions) model.ServiceOptions {
	merged := existing
	if new.Restart != "" {
		merged.Restart = new.Restart
	}
	if new.TimeoutStartSec != "" {
		merged.TimeoutStartSec = new.TimeoutStartSec
	}
	if new.TimeoutStopSec != "" {
		merged.TimeoutStopSec = new.TimeoutStopSec
	}
	merged.Environment = mergeSlices(merged.Environment, new.Environment)
	merged.ExecStartPre = mergeSlices(merged.ExecStartPre, new.ExecStartPre)
	merged.ExecStartPost = mergeSlices(merged.ExecStartPost, new.ExecStartPost)
	return merged
}

func mergeContainerOptions(existing, new model.ContainerOptions) model.ContainerOptions {
	merged := existing
	if new.Image != "" {
		merged.Image = new.Image
	}
	if new.ContainerName != "" {
		merged.ContainerName = new.ContainerName
	}
	merged.Network = mergeSlices(merged.Network, new.Network)
	merged.PublishPort = mergeSlices(merged.PublishPort, new.PublishPort)
	merged.ExposeHostPort = mergeSlices(merged.ExposeHostPort, new.ExposeHostPort)
	merged.Volume = mergeSlices(merged.Volume, new.Volume)
	merged.Mount = mergeSlices(merged.Mount, new.Mount)
	merged.Environment = mergeSlices(merged.Environment, new.Environment)
	merged.EnvironmentFile = mergeSlices(merged.EnvironmentFile, new.EnvironmentFile)
	merged.Secret = mergeSlices(merged.Secret, new.Secret)
	if new.Exec != "" {
		merged.Exec = new.Exec
	}
	if new.Args != "" {
		merged.Args = new.Args
	}
	if new.Entrypoint != "" {
		merged.Entrypoint = new.Entrypoint
	}
	if new.AutoUpdate != "" {
		merged.AutoUpdate = new.AutoUpdate
	}
	if new.Removable != nil {
		merged.Removable = new.Removable
	}
	merged.Label = mergeSlices(merged.Label, new.Label)
	merged.Annotation = mergeSlices(merged.Annotation, new.Annotation)
	if new.User != "" {
		merged.User = new.User
	}
	if new.UserNS != "" {
		merged.UserNS = new.UserNS
	}
	merged.DropCapability = mergeSlices(merged.DropCapability, new.DropCapability)
	merged.AddCapability = mergeSlices(merged.AddCapability, new.AddCapability)
	if new.SecurityLabelDisable != nil {
		merged.SecurityLabelDisable = new.SecurityLabelDisable
	}
	if new.HealthCmd != "" {
		merged.HealthCmd = new.HealthCmd
	}
	if new.Timezone != "" {
		merged.Timezone = new.Timezone
	}
	if new.Pod != "" {
		merged.Pod = new.Pod
	}
	return merged
}

func mergeInstallOptions(existing, new model.InstallOptions) model.InstallOptions {
	merged := existing
	merged.WantedBy = mergeSlices(merged.WantedBy, new.WantedBy)
	merged.RequiredBy = mergeSlices(merged.RequiredBy, new.RequiredBy)
	return merged
}

// mergeSlices returns a union of two string slices, avoiding duplicates.
func mergeSlices(existing, new []string) []string {
	if len(new) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return new
	}

	// Create a map of existing values for O(1) lookups
	seen := make(map[string]bool)
	result := make([]string, 0, len(existing)+len(new))

	for _, v := range existing {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	for _, v := range new {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

func (s *QuadletService) Delete(ctx context.Context, name string) error {
	filePath := filepath.Join(s.dir, name+".container")

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return ErrQuadletNotFound
		}
		return fmt.Errorf("failed to access quadlet file: %w", err)
	}

	// Remove the quadlet file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete quadlet file: %w", err)
	}

	return s.Reload(ctx)
}

func (s *QuadletService) Reload(ctx context.Context) error {
	return s.systemd.ReloadDaemon(ctx)
}
