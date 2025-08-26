package errs

import "errors"

var (
	ErrUnknownType    = errors.New("unknown event type")
	ErrContextTimeout = errors.New("context timeout")
	ErrValidation     = errors.New("validation error")
	ErrNotFound       = errors.New("not found")
)
