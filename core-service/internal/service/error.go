package service

import "errors"

var (
	ErrNotFound = errors.New("not found")

	ErrServiceUnavailable = errors.New("service unavailable")

	ErrBadRequest = errors.New("bad request")

	ErrInternalServiceError = errors.New("internal service error")
)
