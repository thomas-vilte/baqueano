package agent

import (
	"encoding/json"

	"github.com/thomas-vilte/baqueano/internal/tools"
)

// Role identifies who sent a message. The system prompt belongs to Request.
type Role string

const (
	// RoleUser includes questions and tool results.
	RoleUser Role = "user"

	// RoleAssistant includes model text and tool calls.
	RoleAssistant Role = "assistant"
)

// ContentType identifies the payload in a content block.
type ContentType string

const (
	// ContentText selects ContentBlock.Text.
	ContentText ContentType = "text"

	// ContentToolCall selects ContentBlock.ToolCall.
	ContentToolCall ContentType = "tool_call"

	// ContentToolResult selects ContentBlock.ToolResult.
	ContentToolResult ContentType = "tool_result"
)

// Message preserves the order of text, tool calls and tool results.
type Message struct {
	Role    Role           `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock holds only the payload selected by Type.
// Tool-call and tool-result blocks must have a non-nil payload.
type ContentBlock struct {
	Type       ContentType `json:"type"`
	Text       string      `json:"text,omitempty"`
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// ToolCall identifies one invocation, even when several calls use the same tool.
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult answers the invocation identified by CallID.
// On success, Result holds the documents and Text holds the run's status note.
// On failure, Result is nil, Text explains the error and IsError is true.
// Model adapters encode Result followed by Text as the tool-result content.
type ToolResult struct {
	CallID  string        `json:"call_id"`
	Result  *tools.Result `json:"result,omitempty"`
	Text    string        `json:"text,omitempty"`
	IsError bool          `json:"is_error"`
}
