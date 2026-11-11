package agent_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/thomas-vilte/baqueano/internal/agent"
)

func TestErrBudgetExhausted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "direct", err: agent.ErrBudgetExhausted, want: true},
		{
			name: "wrapped",
			err:  fmt.Errorf("agent: force report: %w", agent.ErrBudgetExhausted),
			want: true,
		},
		{name: "other error", err: io.EOF},
		{name: "no error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := errors.Is(tt.err, agent.ErrBudgetExhausted); got != tt.want {
				t.Errorf("errors.Is(err, ErrBudgetExhausted) = %t, want %t", got, tt.want)
			}
		})
	}
}
