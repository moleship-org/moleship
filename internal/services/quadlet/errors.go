package quadlet

import "errors"

var (
	ErrInvalidName = errors.New("invalid unit name")

	ErrInvalidKind = errors.New("invalid quadlet kind")

	ErrInvalidUnit = errors.New("invalid quadlet unit definition")

	ErrUnitNotFound = errors.New("quadlet unit not found")

	ErrUnitAlreadyExists = errors.New("quadlet unit already exists")

	ErrWriteFailed = errors.New("failed to write quadlet unit")

	ErrReadFailed = errors.New("failed to read quadlet unit")

	ErrDeleteFailed = errors.New("failed to delete quadlet unit")

	ErrListFailed = errors.New("failed to list quadlet units")
)
