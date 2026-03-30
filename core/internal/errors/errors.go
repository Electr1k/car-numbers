package errors

import (
	"errors"
	"fmt"
)

type ErrorType string

const (
	ValidationError ErrorType = "validation_error"
	ProviderError ErrorType = "provider_error"
	DatabaseError ErrorType = "database_error"
	InternalError ErrorType = "internal_error"
)

type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:    ValidationError,
		Message: message,
		Err:     err,
	}
}

func NewProviderError(message string, err error) *AppError {
	return &AppError{
		Type:    ProviderError,
		Message: message,
		Err:     err,
	}
}

func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Type:    DatabaseError,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    InternalError,
		Message: message,
		Err:     err,
	}
}

var (
	ErrProviderUnavailable = errors.New("провайдер недоступен")
	ErrProviderInvalidData = errors.New("провайдер вернул невалидные данные")
	ErrUnknownProvider = errors.New("неизвестный провайдер")
	ErrDatabaseOperation = errors.New("ошибка операции с базой данных")
	ErrValidation = errors.New("ошибка валидации данных")
)
