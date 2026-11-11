package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/thomas-vilte/baqueano/internal/tools"
)

func TestDocumentJSONPreservesOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		source     string
		provenance string
	}{
		{
			name:       "built-in source",
			input:      `{"ref":"hn:42","thread":"hn:1","source":"hackernews","provenance":"fetched","author":"alex","snippet":"Synthetic report."}`,
			source:     "hackernews",
			provenance: "fetched",
		},
		{
			name:       "manual input",
			input:      `{"ref":"manual:1","thread":"manual:1","source":"manual","provenance":"provided","snippet":"Synthetic notes."}`,
			source:     "manual",
			provenance: "provided",
		},
		{
			name:       "external source",
			input:      `{"ref":"ext:myforum:42","thread":"ext:myforum:1","source":"myforum","provenance":"external","author":"alex","snippet":"Synthetic report."}`,
			source:     "myforum",
			provenance: "external",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var document tools.Document
			if err := json.Unmarshal([]byte(tt.input), &document); err != nil {
				t.Fatalf("decode document: %v", err)
			}

			if document.Source != tt.source {
				t.Errorf("Source = %q; want %q", document.Source, tt.source)
			}
			if string(document.Provenance) != tt.provenance {
				t.Errorf("Provenance = %q; want %q", document.Provenance, tt.provenance)
			}

			encoded, err := json.Marshal(document)
			if err != nil {
				t.Fatalf("encode document: %v", err)
			}

			var output map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &output); err != nil {
				t.Fatalf("decode output: %v", err)
			}

			for key, want := range map[string]string{
				"source":     tt.source,
				"provenance": tt.provenance,
			} {
				raw, ok := output[key]
				if !ok {
					t.Errorf("JSON output is missing %q", key)
					continue
				}

				var got string
				if err := json.Unmarshal(raw, &got); err != nil {
					t.Fatalf("decode JSON field %q: %v", key, err)
				}
				if got != want {
					t.Errorf("JSON %s = %q; want %q", key, got, want)
				}
			}
		})
	}
}
