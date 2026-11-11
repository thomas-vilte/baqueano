package report

import "github.com/thomas-vilte/baqueano/internal/tools"

// Report contains the findings for one research question.
// Problems and Existing use non-nil slices, including when there are no findings.
type Report struct {
	Question string     `json:"question"`
	Summary  string     `json:"summary"`
	Problems []Problem  `json:"problems"`
	Existing []Existing `json:"existing"`
	Gaps     string     `json:"gaps"`

	// Sources contains cited documents resolved from the run's document index.
	// The submit_report schema must exclude this application-computed field.
	Sources []tools.Document `json:"sources,omitempty"`
}

// Problem describes a reported difficulty and the evidence supporting it.
type Problem struct {
	Title       string     `json:"title"`
	About       string     `json:"about,omitempty"`
	Description string     `json:"description"`
	Workarounds string     `json:"workarounds,omitempty"`
	Evidence    []Evidence `json:"evidence"`

	// Mentions is nil until the application computes the counts.
	// The submit_report schema must exclude this application-computed field.
	Mentions *Mentions `json:"mentions,omitempty"`
}

// Evidence connects a claim to an excerpt from a referenced document.
type Evidence struct {
	Ref string `json:"ref"`

	// Quote is copied from the document and contains at most 30 words.
	Quote string `json:"quote"`
}

// Existing describes a solution or product supported by cited documents.
type Existing struct {
	Text string   `json:"text"`
	Refs []string `json:"refs"`
}

// Mentions counts distinct evidence documents, authors and threads.
type Mentions struct {
	Documents int `json:"documents"`

	// Authors distinguishes the same username on different sources.
	Authors int `json:"authors"`

	Threads int `json:"threads"`

	// FromThreadAuthor counts distinct evidence documents with tools.FlagThreadAuthor.
	FromThreadAuthor int `json:"from_thread_author"`
}
