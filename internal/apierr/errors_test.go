package apierr_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/thomas-vilte/baqueano/internal/apierr"
)

func TestFromStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     int
		success  bool
		category error
	}{
		{name: "before success", code: 199},
		{name: "ok", code: 200, success: true},
		{name: "no content", code: 204, success: true},
		{name: "last success", code: 299, success: true},
		{name: "bad request", code: 400, category: apierr.ErrBadRequest},
		{name: "unauthorized", code: 401, category: apierr.ErrAuth},
		{name: "forbidden", code: 403, category: apierr.ErrAuth},
		{name: "rate limited", code: 429, category: apierr.ErrRateLimited},
		{name: "server error", code: 500, category: apierr.ErrUnavailable},
		{name: "unavailable", code: 503, category: apierr.ErrUnavailable},
		{name: "overloaded", code: 529, category: apierr.ErrUnavailable},
		{name: "last server error", code: 599, category: apierr.ErrUnavailable},
		{name: "first redirect", code: 300},
		{name: "redirect", code: 302},
		{name: "not found", code: 404},
		{name: "before server errors", code: 499},
		{name: "after server errors", code: 600},
		{name: "missing status", code: 0},
	}

	categories := []error{
		apierr.ErrAuth,
		apierr.ErrRateLimited,
		apierr.ErrUnavailable,
		apierr.ErrBadRequest,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, cause := range []error{nil, io.ErrUnexpectedEOF} {
				got := apierr.FromStatus(tt.code, cause)
				if tt.success {
					if got != cause {
						t.Errorf(
							"FromStatus(%d, %v) = %v; want original cause",
							tt.code, cause, got,
						)
					}
					continue
				}

				if got == nil {
					t.Fatalf(
						"FromStatus(%d, %v) = nil; want error",
						tt.code, cause,
					)
				}

				wrapped := fmt.Errorf("read source: %w", got)
				for _, category := range categories {
					want := category == tt.category
					if matches := errors.Is(wrapped, category); matches != want {
						t.Errorf(
							"FromStatus(%d, %v): errors.Is(%v) = %t; want %t",
							tt.code, cause, category, matches, want,
						)
					}
				}

				if cause != nil && !errors.Is(wrapped, cause) {
					t.Errorf(
						"FromStatus(%d, %v) lost the original cause: %v",
						tt.code, cause, got,
					)
				}
			}
		})
	}
}
