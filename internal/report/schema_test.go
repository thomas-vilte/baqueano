package report_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/thomas-vilte/baqueano/internal/report"
)

func TestSubmitSchema(t *testing.T) {
	t.Parallel()
	schema := compileSubmitSchema(t)

	type testCase struct {
		name       string
		target     string
		key        string
		value      any
		remove     bool
		noFindings bool
		nullRoot   bool
		valid      bool
	}
	tests := []testCase{
		{name: "valid report", valid: true},
		{name: "comparison", target: "problem", key: "about", value: "Synthetic backend", valid: true},
		{name: "workaround", target: "problem", key: "workarounds", value: "Teams schedule deployments.", valid: true},
		{name: "empty comparison", target: "problem", key: "about", value: ""},
		{name: "empty workaround", target: "problem", key: "workarounds", value: ""},
		{name: "no findings", noFindings: true, valid: true},
		{name: "null root", nullRoot: true},
		{name: "empty evidence", target: "problem", key: "evidence", value: []any{}},
		{name: "null evidence", target: "problem", key: "evidence", value: nil},
		{name: "empty ref", target: "evidence", key: "ref", value: ""},
		{name: "empty quote", target: "evidence", key: "quote", value: ""},
		{name: "empty existing refs", target: "existing", key: "refs", value: []any{}},
		{name: "empty existing ref", target: "existing", key: "refs", value: []any{""}},
		{name: "empty gaps", key: "gaps", value: ""},
		{name: "question type", key: "question", value: 42},
		{name: "problems type", key: "problems", value: "none"},
		{name: "existing type", key: "existing", value: nil},
		{name: "quote type", target: "evidence", key: "quote", value: 42},
		{name: "refs type", target: "existing", key: "refs", value: "hn:1"},
		{name: "sources supplied", key: "sources", value: []any{}},
		{name: "mentions supplied", target: "problem", key: "mentions", value: map[string]any{}},
		{name: "extra evidence field", target: "evidence", key: "extra", value: true},
		{name: "extra existing field", target: "existing", key: "extra", value: true},
		{name: "legacy problem refs", target: "problem", key: "refs", value: []any{"hn:42"}},
	}
	for _, group := range []struct {
		target string
		fields []string
	}{
		{fields: []string{"question", "summary", "problems", "existing", "gaps"}},
		{target: "problem", fields: []string{"title", "description", "evidence"}},
		{target: "evidence", fields: []string{"ref", "quote"}},
		{target: "existing", fields: []string{"text", "refs"}},
	} {
		for _, field := range group.fields {
			tests = append(tests, testCase{
				name:   "missing " + group.target + "/" + field,
				target: group.target,
				key:    field,
				remove: true,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := submitExample(t)
			problem := input["problems"].([]any)[0].(map[string]any)
			targets := map[string]map[string]any{
				"":         input,
				"problem":  problem,
				"evidence": problem["evidence"].([]any)[0].(map[string]any),
				"existing": input["existing"].([]any)[0].(map[string]any),
			}
			if tt.noFindings {
				input["problems"] = []any{}
				input["existing"] = []any{}
			}
			if tt.key != "" {
				target, ok := targets[tt.target]
				if !ok {
					t.Fatalf("unknown test target %q", tt.target)
				}
				if tt.remove {
					delete(target, tt.key)
				} else {
					target[tt.key] = tt.value
				}
			}
			var instance any = input
			if tt.nullRoot {
				instance = nil
			}

			err := schema.Validate(instance)
			if tt.valid {
				if err != nil {
					t.Fatalf("valid report rejected: %v", err)
				}
				return
			}
			if _, ok := errors.AsType[*jsonschema.ValidationError](err); !ok {
				t.Fatalf("Validate = %v; want a schema validation error", err)
			}
		})
	}
}

func TestSubmitSchemaReturnsIndependentBytes(t *testing.T) {
	t.Parallel()

	first := report.SubmitSchema()
	second := report.SubmitSchema()
	original := bytes.Clone(first)
	if !json.Valid(first) || !bytes.Equal(first, second) {
		t.Fatal("SubmitSchema must return stable, valid JSON")
	}
	first[0] = '['
	if !bytes.Equal(second, original) {
		t.Fatal("SubmitSchema calls share their bytes")
	}
	if got := report.SubmitSchema(); !bytes.Equal(got, original) {
		t.Fatal("mutating one schema changed a later result")
	}
}

func compileSubmitSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	var document any
	if err := json.Unmarshal(report.SubmitSchema(), &document); err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(nil)
	compiler.DefaultDraft(jsonschema.Draft2020)

	const location = "https://baqueano.invalid/submit_report.schema.json"
	if err := compiler.AddResource(location, document); err != nil {
		t.Fatalf("register schema: %v", err)
	}
	schema, err := compiler.Compile(location)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return schema
}

func submitExample(t *testing.T) map[string]any {
	t.Helper()

	const input = `{
		"question": "What problems do users report with shared state?",
		"summary": "Teams report that a shared lock delays deployments.",
		"problems": [{
			"title": "Deployments wait for a shared lock",
			"description": "Teams cannot deploy while another team holds the lock.",
			"evidence": [{
				"ref": "hn:42",
				"quote": "Deployments wait for another team to release the shared state lock."
			}]
		}],
		"existing": [{
			"text": "A synthetic service coordinates deployments.",
			"refs": ["hn:1"]
		}],
		"gaps": "The synthetic sample only covers one source."
	}`
	var document map[string]any
	if err := json.Unmarshal([]byte(input), &document); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return document
}
