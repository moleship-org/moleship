package quadlet

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/moleship-org/moleship/internal/config"
)

const (
	dirPerm = 0o755

	filePerm = 0o644

	rootDir = "/etc/containers/systemd"
)

type NewFilesystemParams struct {
	BaseDir string

	// rootless -> $XDG_CONFIG_HOME/containers/systemd (or
	// ~/.config/containers/systemd), root -> /etc/containers/systemd.
	UserMode bool
}

type Filesystem struct {
	dir string
}

var _ FSPort = (*Filesystem)(nil)

func NewFilesystem(params *NewFilesystemParams) (*Filesystem, error) {
	if params == nil {
		params = new(NewFilesystemParams)
	}

	dir := params.BaseDir
	if dir == "" {
		resolved, err := defaultDir(params.UserMode)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve default quadlet dir: %w", err)
		}
		dir = resolved
	}

	return &Filesystem{dir: dir}, nil
}

func (f *Filesystem) Dir() string {
	return f.dir
}

func (f *Filesystem) pathFor(kind Kind, name string) (string, error) {
	if !kind.Valid() {
		return "", ErrInvalidKind
	}
	if err := validateName(name); err != nil {
		return "", err
	}
	filename := name + "." + kind.Extension()
	return filepath.Join(f.dir, filename), nil
}

func (f *Filesystem) Write(ctx context.Context, unit Unit, opts WriteOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	kind := unit.Kind()
	name := unit.Name()

	dest, err := f.pathFor(kind, name)
	if err != nil {
		return err
	}

	if opts.FailIfExists {
		if _, statErr := os.Stat(dest); statErr == nil {
			return ErrUnitAlreadyExists
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			return fmt.Errorf("%w: %v", ErrWriteFailed, statErr)
		}
	}

	content, err := unit.Render()
	if err != nil {
		return fmt.Errorf("%w: failed to render unit: %v", ErrWriteFailed, err)
	}

	if err := os.MkdirAll(f.dir, dirPerm); err != nil {
		return fmt.Errorf("%w: failed to create dir: %v", ErrWriteFailed, err)
	}

	tmp, err := os.CreateTemp(f.dir, "."+name+".*.tmp")
	if err != nil {
		return fmt.Errorf("%w: failed to create temp file: %v", ErrWriteFailed, err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return fmt.Errorf("%w: failed to write temp file: %v", ErrWriteFailed, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("%w: failed to sync temp file: %v", ErrWriteFailed, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("%w: failed to close temp file: %v", ErrWriteFailed, err)
	}
	if err := os.Chmod(tmpPath, filePerm); err != nil {
		return fmt.Errorf("%w: failed to set permissions: %v", ErrWriteFailed, err)
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		return fmt.Errorf("%w: failed to rename into place: %v", ErrWriteFailed, err)
	}

	return nil
}

func (f *Filesystem) Read(ctx context.Context, kind Kind, name string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	path, err := f.pathFor(kind, name)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", ErrUnitNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	return string(data), nil
}

func (f *Filesystem) Delete(ctx context.Context, kind Kind, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := f.pathFor(kind, name)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrUnitNotFound
		}
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	return nil
}

func (f *Filesystem) List(ctx context.Context) ([]Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(f.dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrListFailed, err)
	}

	extToKind := make(map[string]Kind, len(AllKinds))
	for _, k := range AllKinds {
		extToKind[k.Extension()] = k
	}

	result := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := strings.TrimPrefix(filepath.Ext(e.Name()), ".")
		kind, ok := extToKind[ext]
		if !ok {
			continue
		}

		name := strings.TrimSuffix(e.Name(), "."+ext)
		result = append(result, Entry{Name: name, Kind: kind})
	}

	return result, nil
}

func defaultDir(userMode bool) (string, error) {
	if !userMode {
		return rootDir, nil
	}
	return filepath.Join(string(os.PathSeparator), "home", config.HOST_USER, ".config", "containers", "systemd"), nil
}

func validateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	if name != filepath.Base(name) {
		return ErrInvalidName
	}
	if strings.Contains(name, "..") {
		return ErrInvalidName
	}
	return nil
}
