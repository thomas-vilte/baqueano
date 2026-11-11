package report_test

import (
	"encoding/json"
	"maps"
	"testing"

	"github.com/thomas-vilte/baqueano/internal/report"
	"github.com/thomas-vilte/baqueano/internal/tools"
)

func TestReportJSONComputedFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mentions   *report.Mentions
		sources    []tools.Document
		wantCounts map[string]int
	}{
		{name: "not computed"},
		{name: "empty sources", sources: []tools.Document{}},
		{
			name:     "computed zeros",
			mentions: &report.Mentions{},
			wantCounts: map[string]int{
				"documents":          0,
				"authors":            0,
				"threads":            0,
				"from_thread_author": 0,
			},
		},
		{
			name: "computed counts and cited source",
			mentions: &report.Mentions{
				Documents:        4,
				Authors:          3,
				Threads:          2,
				FromThreadAuthor: 1,
			},
			sources: []tools.Document{{
				Ref:        "hn:42",
				Thread:     "hn:1",
				Source:     "hackernews",
				Provenance: tools.ProvenanceFetched,
				Author:     "alex",
				Snippet:    "Synthetic report.",
				Flags:      []tools.Flag{tools.FlagThreadAuthor},
			}},
			wantCounts: map[string]int{
				"documents":          4,
				"authors":            3,
				"threads":            2,
				"from_thread_author": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value := report.Report{
				Problems: []report.Problem{{Mentions: tt.mentions}},
				Existing: []report.Existing{},
				Sources:  tt.sources,
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("encode report: %v", err)
			}

			var output map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &output); err != nil {
				t.Fatalf("decode output: %v", err)
			}
			var problems []map[string]json.RawMessage
			if err := json.Unmarshal(output["problems"], &problems); err != nil {
				t.Fatalf("decode problems: %v", err)
			}
			if len(problems) != 1 {
				t.Fatalf("problems length = %d; want 1", len(problems))
			}

			rawMentions, hasMentions := problems[0]["mentions"]
			if want := tt.mentions != nil; hasMentions != want {
				t.Errorf("mentions present = %t; want %t", hasMentions, want)
			}
			if hasMentions {
				var counts map[string]int
				if err := json.Unmarshal(rawMentions, &counts); err != nil {
					t.Fatalf("decode mentions: %v", err)
				}
				if !maps.Equal(counts, tt.wantCounts) {
					t.Errorf("mentions = %v; want %v", counts, tt.wantCounts)
				}
			}

			rawSources, hasSources := output["sources"]
			if want := len(tt.sources) > 0; hasSources != want {
				t.Errorf("sources present = %t; want %t", hasSources, want)
			}
			if hasSources {
				var sources []map[string]json.RawMessage
				if err := json.Unmarshal(rawSources, &sources); err != nil {
					t.Fatalf("decode sources: %v", err)
				}
				if len(sources) != len(tt.sources) {
					t.Fatalf(
						"sources length = %d; want %d",
						len(sources), len(tt.sources),
					)
				}
				for i, source := range sources {
					var ref string
					if err := json.Unmarshal(source["ref"], &ref); err != nil {
						t.Fatalf("decode source ref: %v", err)
					}
					if ref != tt.sources[i].Ref {
						t.Errorf(
							"source ref = %q; want %q",
							ref, tt.sources[i].Ref,
						)
					}
				}
			}
		})
	}
}
