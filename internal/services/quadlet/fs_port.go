package quadlet

import "context"

type WriteOptions struct {
	FailIfExists bool
}

type Entry struct {
	Name string
	Kind Kind
}

type FSPort interface {
	Write(ctx context.Context, unit Unit, opts WriteOptions) error

	Read(ctx context.Context, kind Kind, name string) (string, error)

	Delete(ctx context.Context, kind Kind, name string) error

	List(ctx context.Context) ([]Entry, error)

	Dir() string
}
