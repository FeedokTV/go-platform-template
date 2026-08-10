package apperror

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
	// ErrInternal: marker for failures already logged at the throw site;
	// Adapters map it to a generic 5xx WITHOUT logging;
	// Domains never wrap this it is boundary plumbing
	ErrInternal = errors.New("internal error")
)
