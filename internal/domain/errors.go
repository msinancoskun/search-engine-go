package domain

import (
	"fmt"
)

type ErrorCode string

const (
	ErrorCodeNotFound ErrorCode = "NOT_FOUND"
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
	ErrorCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrorCodeProviderError ErrorCode = "PROVIDER_ERROR"
	ErrorCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrorCodeCacheError ErrorCode = "CACHE_ERROR"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(resource string, identifier interface{}) *DomainError {
	return &DomainError{
		Code:    ErrorCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]interface{}{
			"resource":   resource,
			"identifier": identifier,
		},
	}
}

func NewInvalidInputError(field string, reason string) *DomainError {
	return &DomainError{
		Code:    ErrorCodeInvalidInput,
		Message: fmt.Sprintf("invalid input for field '%s': %s", field, reason),
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Code:    ErrorCodeInternalError,
		Message: message,
		Err:     err,
	}
}

func NewProviderError(provider string, message string, err error) *DomainError {
	return &DomainError{
		Code:    ErrorCodeProviderError,
		Message: fmt.Sprintf("provider '%s' error: %s", provider, message),
		Details: map[string]interface{}{
			"provider": provider,
		},
		Err: err,
	}
}

func NewDatabaseError(operation string, err error) *DomainError {
	return &DomainError{
		Code:    ErrorCodeDatabaseError,
		Message: fmt.Sprintf("database error during %s", operation),
		Details: map[string]interface{}{
			"operation": operation,
		},
		Err: err,
	}
}

func NewCacheError(operation string, err error) *DomainError {
	return &DomainError{
		Code:    ErrorCodeCacheError,
		Message: fmt.Sprintf("cache error during %s", operation),
		Details: map[string]interface{}{
			"operation": operation,
		},
		Err: err,
	}
}

func IsNotFoundError(err error) bool {
	var domainErr *DomainError
	if err != nil && err.Error() != "" {
		if de, ok := err.(*DomainError); ok {
			domainErr = de
		}
	}
	return domainErr != nil && domainErr.Code == ErrorCodeNotFound
}

func IsInvalidInputError(err error) bool {
	var domainErr *DomainError
	if err != nil && err.Error() != "" {
		if de, ok := err.(*DomainError); ok {
			domainErr = de
		}
	}
	return domainErr != nil && domainErr.Code == ErrorCodeInvalidInput
}
