package agent_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/thomas-vilte/baqueano/internal/agent"
)

func TestResponseToolCalls(t *testing.T) {
	t.Parallel()

	first := agent.ToolCall{
		ID: "call-1", Name: "search_hackernews",
		Input: json.RawMessage(`{"query":"terraform state"}`),
	}
	second := agent.ToolCall{
		ID: "call-2", Name: "search_hackernews",
		Input: json.RawMessage(`{"query":"terraform locking"}`),
	}

	tests := []struct {
		name    string
		message agent.Message
		want    []agent.ToolCall
	}{
		{
			name: "empty response",
			want: []agent.ToolCall{},
		},
		{
			name: "text only",
			message: agent.Message{
				Role: agent.RoleAssistant,
				Content: []agent.ContentBlock{
					{Type: agent.ContentText, Text: "Finished."},
				},
			},
			want: []agent.ToolCall{},
		},
		{
			name: "calls keep their order and IDs",
			message: agent.Message{
				Role: agent.RoleAssistant,
				Content: []agent.ContentBlock{
					{Type: agent.ContentText, Text: "I'll search both terms."},
					{Type: agent.ContentToolCall, ToolCall: &first},
					{Type: agent.ContentText, Text: "I'll also check locking."},
					{Type: agent.ContentToolCall, ToolCall: &second},
				},
			},
			want: []agent.ToolCall{first, second},
		},
		{
			name: "tool results are not calls",
			message: agent.Message{
				Role: agent.RoleUser,
				Content: []agent.ContentBlock{
					{
						Type: agent.ContentToolResult,
						ToolResult: &agent.ToolResult{
							CallID: "call-1", Text: "Source unavailable.", IsError: true,
						},
					},
				},
			},
			want: []agent.ToolCall{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			response := agent.Response{Message: tt.message}
			if got := response.ToolCalls(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToolCalls() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestResponseToolCallsCopiesInput(t *testing.T) {
	t.Parallel()

	input := json.RawMessage(`{"query":"terraform"}`)
	response := agent.Response{
		Message: agent.Message{
			Role: agent.RoleAssistant,
			Content: []agent.ContentBlock{
				{
					Type: agent.ContentToolCall,
					ToolCall: &agent.ToolCall{
						ID: "call-1", Name: "search_hackernews", Input: input,
					},
				},
			},
		},
	}

	calls := response.ToolCalls()
	if len(calls) != 1 {
		t.Fatalf("len(ToolCalls()) = %d, want 1", len(calls))
	}

	calls[0].ID = "changed"
	calls[0].Name = "changed"
	calls[0].Input[0] = '['

	want := []agent.ToolCall{
		{
			ID: "call-1", Name: "search_hackernews",
			Input: json.RawMessage(`{"query":"terraform"}`),
		},
	}
	if got := response.ToolCalls(); !reflect.DeepEqual(got, want) {
		t.Errorf("mutating returned calls changed the response: got %#v, want %#v", got, want)
	}
}
