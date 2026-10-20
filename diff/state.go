package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type state struct {
	graph schemaGraph

	// The name of each document's components.schemas entries.
	baseComponents     map[*openapi3.Schema]string
	revisionComponents map[*openapi3.Schema]string
}

func newState() *state {
	return &state{graph: newSchemaGraph()}
}
