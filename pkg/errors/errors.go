package apperrors

import "errors"
var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("resource already exists")
	ErrValidation     = errors.New("validation failed")
	ErrInternalServer = errors.New("internal server error")
	ErrBadRequest     = errors.New("bad request")
	ErrTooManyRequest = errors.New("too many requests")
)
