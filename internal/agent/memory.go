package agent

import (
	"context"

	"github.com/thomas-vilte/baqueano/internal/report"
)

// Memory owns one run's working context and any session history passed to it.
// The run calls its methods sequentially; concurrent runs use separate instances.
type Memory interface {
	// Messages returns an independent snapshot within maxTokens.
	// Compaction may call a model and must respect ctx.
	// It must preserve document refs and tool-call/result pairs.
	Messages(ctx context.Context, maxTokens int) ([]Message, error)

	// Append stores an independent copy, including nested slices and JSON input.
	Append(message Message)

	// EndTurn retains a copy of the question and report for session history
	// and releases the working messages.
	EndTurn(result report.Report)
}
