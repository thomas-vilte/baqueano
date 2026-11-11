package agent

import "time"

// Budget holds limits for one research question.
// Counters and elapsed time belong to the run, not to this value.
type Budget struct {
	MaxSteps int

	// MaxInputTokens limits each model request, not cumulative usage.
	MaxInputTokens int

	Timeout     time.Duration
	ToolTimeout time.Duration

	// MaxCostUSD limits the accumulated estimated cost of the run.
	MaxCostUSD float64
}

type sentinel string

func (e sentinel) Error() string {
	return string(e)
}

// ErrBudgetExhausted tells the loop to force a report.
// A report produced after exhaustion carries StatusBudgetExhausted;
// exhaustion alone is not an error returned alongside that report.
const ErrBudgetExhausted sentinel = "budget exhausted"
