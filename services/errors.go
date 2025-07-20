package services

import (
	"fmt"
	"net/http"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeAPI        ErrorType = "api"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeAuth       ErrorType = "auth"
	ErrorTypeConfig     ErrorType = "config"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeRateLimit  ErrorType = "rate_limit"
	ErrorTypeNotFound   ErrorType = "not_found"
)

// GitHubError represents a structured error with context
type GitHubError struct {
	Type       ErrorType
	Message    string
	StatusCode int
	Repo       string
	Underlying error
}

func (e *GitHubError) Error() string {
	if e.Repo != "" {
		return fmt.Sprintf("[%s] %s (repo: %s)", e.Type, e.Message, e.Repo)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *GitHubError) Unwrap() error {
	return e.Underlying
}

// IsRetryable returns true if the error is potentially recoverable with a retry
func (e *GitHubError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeNetwork, ErrorTypeTimeout:
		return true
	case ErrorTypeAPI:
		// 5xx errors are retryable, 4xx generally are not
		return e.StatusCode >= 500
	case ErrorTypeRateLimit:
		return true
	default:
		return false
	}
}

// NewAPIError creates a new API-related error
func NewAPIError(message string, statusCode int, repo string, underlying error) *GitHubError {
	errorType := ErrorTypeAPI
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		errorType = ErrorTypeAuth
	case http.StatusTooManyRequests:
		errorType = ErrorTypeRateLimit
	case http.StatusNotFound:
		errorType = ErrorTypeNotFound
	case http.StatusBadRequest:
		errorType = ErrorTypeValidation
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		errorType = ErrorTypeNetwork
	case http.StatusGatewayTimeout:
		errorType = ErrorTypeTimeout
	default:
		if statusCode >= 500 {
			errorType = ErrorTypeNetwork
		}
	}

	return &GitHubError{
		Type:       errorType,
		Message:    message,
		StatusCode: statusCode,
		Repo:       repo,
		Underlying: underlying,
	}
}

// NewNetworkError creates a new network-related error
func NewNetworkError(message string, repo string, underlying error) *GitHubError {
	return &GitHubError{
		Type:       ErrorTypeNetwork,
		Message:    message,
		Repo:       repo,
		Underlying: underlying,
	}
}

// NewTimeoutError creates a new timeout-related error
func NewTimeoutError(message string, repo string, underlying error) *GitHubError {
	return &GitHubError{
		Type:       ErrorTypeTimeout,
		Message:    message,
		Repo:       repo,
		Underlying: underlying,
	}
}

// NewConfigError creates a new configuration-related error
func NewConfigError(message string, underlying error) *GitHubError {
	return &GitHubError{
		Type:       ErrorTypeConfig,
		Message:    message,
		Underlying: underlying,
	}
}

// NewValidationError creates a new validation-related error
func NewValidationError(message string, repo string, underlying error) *GitHubError {
	return &GitHubError{
		Type:       ErrorTypeValidation,
		Message:    message,
		Repo:       repo,
		Underlying: underlying,
	}
}
