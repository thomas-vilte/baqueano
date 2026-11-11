package agent

import (
	"bytes"
	"context"

	"github.com/thomas-vilte/baqueano/internal/tools"
)

// Model translates between these contracts and a provider's API.
// Implementations must respect cancellation and support concurrent calls.
// Complete must not mutate or retain request data after returning.
// Responses must not share buffers that the adapter later reuses.
type Model interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

// Request contains one model call, independent of the provider's SDK.
type Request struct {
	System          string             `json:"system"`
	Messages        []Message          `json:"messages"`
	Tools           []tools.Definition `json:"tools"`
	ToolChoice      ToolChoice         `json:"tool_choice"`
	MaxOutputTokens int                `json:"max_output_tokens"`
}

// ToolChoice forces the named tool. Its zero value lets the model choose.
// A forced tool must be present in Request.Tools.
type ToolChoice struct {
	Name string `json:"name,omitempty"`
}

// StopReason describes why the model stopped generating.
// The zero value is unknown; adapters must reject unsupported reasons.
type StopReason string

const (
	// StopEndTurn means the model finished its turn.
	StopEndTurn StopReason = "end_turn"

	// StopToolUse means the model requested one or more tools.
	StopToolUse StopReason = "tool_use"

	// StopMaxTokens means the response reached its output token limit.
	StopMaxTokens StopReason = "max_tokens"
)

// Usage records token counts and the adapter's estimated cost in USD.
// Response.Usage covers one model call; Result.Usage accumulates a run's calls.
type Usage struct {
	// InputTokens includes cached input; cache counters below are subsets.
	InputTokens              int     `json:"input_tokens"`
	OutputTokens             int     `json:"output_tokens"`
	CacheReadInputTokens     int     `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int     `json:"cache_creation_input_tokens"`
	CostUSD                  float64 `json:"cost_usd"`
}

// Response keeps the model's message intact for memory and tool execution.
type Response struct {
	// ID and Model come from the provider when available.
	ID    string `json:"id,omitempty"`
	Model string `json:"model,omitempty"`

	Message    Message    `json:"message"`
	StopReason StopReason `json:"stop_reason"`
	Usage      Usage      `json:"usage"`
}

// ToolCalls returns independent copies of the calls in message order.
// It returns a non-nil empty slice when there are no calls.
// Adapters validate content blocks before returning a Response.
func (r *Response) ToolCalls() []ToolCall {
	calls := make([]ToolCall, 0)

	for _, block := range r.Message.Content {
		if block.Type != ContentToolCall || block.ToolCall == nil {
			continue
		}

		call := *block.ToolCall
		call.Input = bytes.Clone(call.Input)
		calls = append(calls, call)
	}

	return calls
}
