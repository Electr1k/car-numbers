package service

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")

	ErrServiceUnavailable = errors.New("service unavailable")

	ErrBadRequest = errors.New("bad request")

	ErrUnprocessable = errors.New("unprocessable")

	ErrInternalServiceError = errors.New("internal service error")
)

type ClientError struct {
	Err     error
	Code    string
	Message string
}

func NewClientError(err error, code string, message string) *ClientError {
	return &ClientError{Err: err, Code: code, Message: message}
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("%v: %s: %s", e.Err, e.Code, e.Message)
}

func (e *ClientError) Unwrap() error {
	return e.Err
}
