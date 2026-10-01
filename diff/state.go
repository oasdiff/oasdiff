package diff

import "github.com/getkin/kin-openapi/openapi3"

type state struct {
	graph schemaGraph

	// The name of each document's components.schemas entries, empty when the
	// comparison was started without documents.
	baseComponents     map[*openapi3.Schema]string
	revisionComponents map[*openapi3.Schema]string
}

func newState() *state {
	return &state{graph: newSchemaGraph()}
}

func componentNames(spec *openapi3.T) map[*openapi3.Schema]string {
	names := map[*openapi3.Schema]string{}
	if spec == nil || spec.Components == nil {
		return names
	}
	for name, schemaRef := range spec.Components.Schemas {
		if schemaRef != nil && schemaRef.Value != nil {
			names[schemaRef.Value] = name
		}
	}
	return names
}
