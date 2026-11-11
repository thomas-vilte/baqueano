package tools

import "time"

// Provenance records how a document entered the research.
type Provenance string

const (
	// ProvenanceFetched identifies documents retrieved by a built-in source.
	ProvenanceFetched Provenance = "fetched"

	// ProvenanceProvided identifies documents supplied by a person.
	ProvenanceProvided Provenance = "provided"

	// ProvenanceExternal identifies documents returned by an external tool.
	ProvenanceExternal Provenance = "external"
)

// Flag marks context that affects how evidence should be read.
type Flag string

const (
	// FlagLaunch marks a Show HN or Launch HN post.
	FlagLaunch Flag = "launch"
	// FlagThreadAuthor marks a comment written by the thread's author.
	FlagThreadAuthor Flag = "thread_author"
)

// Document is a source excerpt with the metadata needed to verify citations.
type Document struct {
	Ref string `json:"ref"`

	// Thread is the root document's ref; it equals Ref for a root document.
	Thread string `json:"thread"`

	Source     string     `json:"source"`
	Provenance Provenance `json:"provenance"`
	Title      string     `json:"title,omitempty"`
	Author     string     `json:"author,omitempty"`
	URL        string     `json:"url,omitempty"`

	// CreatedAt is normalized to UTC by the source; zero means unknown.
	CreatedAt time.Time `json:"created_at,omitzero"`

	Score int `json:"score,omitempty"`

	// Snippet is the received excerpt against which quotes are checked.
	Snippet string `json:"snippet"`

	Flags []Flag `json:"flags,omitempty"`
}
