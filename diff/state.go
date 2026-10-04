package diff

import (
	"strings"

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

// componentName names a compared schema after its components.schemas entry. A
// $ref with fields beside it, such as an OpenAPI 3.1 description override,
// resolves to a copy of the entry that names does not hold, so the copy is
// named from the $ref. A $ref into an entry, rather than to one, names a
// schema inside it and has no name.
func componentName(names map[*openapi3.Schema]string, schema *openapi3.SchemaRef) string {
	if name, ok := names[schema.Value]; ok {
		return name
	}
	name, ok := strings.CutPrefix(schema.Ref, "#/components/schemas/")
	if !ok || strings.Contains(name, "/") {
		return ""
	}
	return name
}
