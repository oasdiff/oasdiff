package diff

import "github.com/getkin/kin-openapi/openapi3"

type state struct {
	graph schemaGraph

	// baseComponents and revisionComponents name the schema each
	// components.schemas entry of a document holds, so a node of the graph
	// can carry the name of the component it compares (SchemaDiff.
	// BaseComponent). Empty for a comparison started without documents, as
	// SchemaRefsValidationEquivalent does.
	baseComponents     map[*openapi3.Schema]string
	revisionComponents map[*openapi3.Schema]string
}

func newState() *state {
	return &state{graph: newSchemaGraph()}
}

// componentNames maps each schema held by a components.schemas entry to its
// name.
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
