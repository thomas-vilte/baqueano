package apierr

import (
	"fmt"
	"net/http"
)

type sentinel string

func (e sentinel) Error() string {
	return string(e)
}

const (
	// ErrAuth indicates missing, invalid or expired credentials.
	ErrAuth sentinel = "authentication failed"

	// ErrRateLimited indicates a rate limit or a provider's backoff.
	ErrRateLimited sentinel = "rate limited"

	// ErrUnavailable indicates a server failure, overload or timeout.
	ErrUnavailable sentinel = "service unavailable"

	// ErrBadRequest indicates an invalid request.
	ErrBadRequest sentinel = "bad request"
)

// FromStatus classifies an HTTP response while preserving err.
// A 2xx status returns err unchanged. Other statuses always return an error,
// even when no category applies.
func FromStatus(code int, err error) error {
	if code >= http.StatusOK && code < http.StatusMultipleChoices {
		return err
	}

	var category error

	switch {
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		category = ErrAuth
	case code == http.StatusTooManyRequests:
		category = ErrRateLimited
	case code >= http.StatusInternalServerError && code < 600:
		category = ErrUnavailable
	case code == http.StatusBadRequest:
		category = ErrBadRequest
	}

	if category == nil {
		if err != nil {
			return fmt.Errorf("http status %d: %w", code, err)
		}
		return fmt.Errorf("http status %d", code)
	}

	if err == nil {
		return fmt.Errorf("http status %d: %w", code, category)
	}

	return fmt.Errorf("http status %d: %w: %w", code, category, err)
}
