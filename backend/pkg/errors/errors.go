package errors

import "fmt"

type ErrorType string

const (
	NotFound      ErrorType = "NOT_FOUND"
	Unauthorized  ErrorType = "UNAUTHORIZED"
	InternalError ErrorType = "INTERNAL_ERROR"
)

type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func New(t ErrorType, msg string, err error) *AppError {
	return &AppError{
		Type:    t,
		Message: msg,
		Err:     err,
	}
}
