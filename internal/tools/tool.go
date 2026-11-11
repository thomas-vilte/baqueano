package tools

import (
	"context"
	"encoding/json"
)

// Tool is a source action available to the model.
// Implementations must respect cancellation and support concurrent calls.
// Name, Description and InputSchema must remain stable after registration.
type Tool interface {
	Name() string
	Description() string
	InputSchema() json.RawMessage
	Call(ctx context.Context, input json.RawMessage) (Result, error)
}

// Definition is the tool metadata sent to a model.
type Definition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// Result contains the documents returned by a tool call.
// Successful calls return a non-nil Documents slice, including for empty results.
type Result struct {
	Documents []Document `json:"documents"`
	Truncated bool       `json:"truncated"`
}
