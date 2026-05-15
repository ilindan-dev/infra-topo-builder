package domain

import "errors"

// Common domain errors returned by the core logic and repository ports.
// These errors should be used for layer-to-layer communication to avoid
// coupling with specific database driver errors.
var (
	// ErrNotFound indicates that the requested entity does not exist in the data store.
	ErrNotFound = errors.New("record not found")

	// ErrAlreadyExists indicates a violation of a uniqueness constraint
	// (e.g., trying to create a node with a GUID that is already registered).
	ErrAlreadyExists = errors.New("record already exists")

	// ErrInvalidData indicates that the provided entity or payload fails
	// business validation rules.
	ErrInvalidData = errors.New("invalid data provided")

	// ErrConflict indicates a state conflict in the underlying data store,
	// such as a foreign key violation.
	ErrConflict = errors.New("data conflict (e.g. foreign key violation)")

	// ErrInternalServer indicates an unexpected, unrecoverable technical failure.
	ErrInternalServer = errors.New("internal server error")

	// ErrInvalidNodeKind indicate an unexpected kind and not support.
	ErrInvalidNodeKind = errors.New("invalid node kind")
)
