package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type state struct {
	graph schemaGraph

	// The name of each document's components.schemas entries.
	baseComponents     map[*openapi3.Schema]string
	revisionComponents map[*openapi3.Schema]string

	// The callback pairs being compared. A callback can reach itself through
	// its operations' callbacks.
	callbacksInProgress map[callbackPair]struct{}
}

type callbackPair struct {
	base, revision *openapi3.Callback
}

func newState() *state {
	return &state{graph: newSchemaGraph(), callbacksInProgress: map[callbackPair]struct{}{}}
}
