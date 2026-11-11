package agent

import "github.com/thomas-vilte/baqueano/internal/report"

// Status describes how a run produced its report.
type Status string

const (
	// StatusComplete means the run finished without exhausting a budget or losing a source.
	StatusComplete Status = "complete"

	// StatusBudgetExhausted means a limit forced the report.
	// It takes precedence over StatusPartial; failed sources remain in Result.
	StatusBudgetExhausted Status = "budget_exhausted"

	// StatusPartial means the report was produced with one or more unavailable sources.
	StatusPartial Status = "partial"
)

// Result carries a report and the metadata collected while producing it.
// A run returns a non-nil error only when it cannot produce a report.
type Result struct {
	Report report.Report `json:"report"`
	Status Status        `json:"status"`

	// FailedSources contains distinct source IDs from the tool registry.
	// Successful runs without failed sources use a non-nil empty slice.
	FailedSources []string `json:"failed_sources"`

	Usage Usage `json:"usage"`

	// Turns counts loop calls to Model.Complete, including a forced report.
	// It excludes retries performed inside the model adapter.
	Turns int `json:"turns"`
}
