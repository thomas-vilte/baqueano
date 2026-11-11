package report

import "encoding/json"

// SubmitSchema returns the input schema for submit_report.
// Each call returns its own bytes. Sources and mention counts are excluded.
func SubmitSchema() json.RawMessage {
	return json.RawMessage(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "submit_report",
  "type": "object",
  "additionalProperties": false,
  "required": ["question", "summary", "problems", "existing", "gaps"],
  "properties": {
    "question": {"type": "string", "minLength": 1},
    "summary": {
      "type": "string",
      "minLength": 1,
      "description": "Summarize the findings in three to five lines."
    },
    "problems": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["title", "description", "evidence"],
        "properties": {
          "title": {"type": "string", "minLength": 1},
          "about": {
            "type": "string",
            "minLength": 1,
            "description": "For comparisons, name the option this problem concerns."
          },
          "description": {"type": "string", "minLength": 1},
          "workarounds": {
            "type": "string",
            "minLength": 1,
            "description": "Include only workarounds mentioned in the documents."
          },
          "evidence": {
            "type": "array",
            "minItems": 1,
            "items": {
              "type": "object",
              "additionalProperties": false,
              "required": ["ref", "quote"],
              "properties": {
                "ref": {"type": "string", "minLength": 1},
                "quote": {
                  "type": "string",
                  "minLength": 1,
                  "description": "Copy at most 30 words from the referenced document."
                }
              }
            }
          }
        }
      }
    },
    "existing": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["text", "refs"],
        "properties": {
          "text": {"type": "string", "minLength": 1},
          "refs": {
            "type": "array",
            "minItems": 1,
            "items": {"type": "string", "minLength": 1}
          }
        }
      }
    },
    "gaps": {
      "type": "string",
      "minLength": 1,
      "description": "Describe missing evidence, failed sources or limits of the research."
    }
  }
}`)
}
