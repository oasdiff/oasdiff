package diff

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ModifiedSubschemas is list of modified subschemas with their diffs
// Unlike other Modiefied types which are modeled as maps, this one is modeled as a slice to avoid complex mapping keys
type ModifiedSubschemas []*ModifiedSubschema

// ModifiedSubschema represents a modified subschema with its indentifiers in base and revision, and the schema diff
type ModifiedSubschema struct {
	Base     Subschema   `json:"base" yaml:"base"`
	Revision Subschema   `json:"revision" yaml:"revision"`
	Diff     *SchemaDiff `json:"diff" yaml:"diff"`
}

// String returns a string representation of the modified subschema
func (modifiedSchema *ModifiedSubschema) String() string {
	baseName := modifiedSchema.Base.String()
	revisonName := modifiedSchema.Revision.String()

	if baseName == revisonName {
		return baseName
	}

	return fmt.Sprintf("%s -> %s", baseName, revisonName)
}

// Subschemas is a list of subschemas
type Subschemas []Subschema

// Subschema uniquely identifies a subschema by its index, component and title
type Subschema struct {
	Index     int    `json:"index" yaml:"index"`                             // zero-based index in the schema's subschemas
	Component string `json:"component,omitempty" yaml:"component,omitempty"` // component name if the subschema is a reference to components/schemas
	Title     string `json:"title,omitempty" yaml:"title,omitempty"`         // title of the subschema
}

// String returns a string representation of the subschema
// Note that we convert the index to 1-based index
func (subschema Subschema) String() string {
	const prefix = "subschema"

	if subschema.Title != "" {
		return fmt.Sprintf("%s #%d: %s", prefix, subschema.Index+1, subschema.Title)
	}

	// note: we may want to ad the index to the component name in the future
	if subschema.Component != "" {
		return fmt.Sprintf("#/components/schemas/%s", subschema.Component)
	}

	return fmt.Sprintf("%s #%d", prefix, subschema.Index+1)
}

func getSubschemas(indexes []int, schemaRefs openapi3.SchemaRefs) Subschemas {
	result := Subschemas{}
	for _, index := range indexes {
		result = append(result, Subschema{
			Index: index,
			Title: schemaRefs[index].Value.Title,
		})
	}
	return result
}

// String returns a string representation of the subschemas
func (schemas Subschemas) String() string {
	names := make([]string, len(schemas))
	for i, schema := range schemas {
		names[i] = schema.String()
	}
	return strings.Join(names, ", ")
}

func (modifiedSchemas ModifiedSubschemas) addSchemaDiff(config *Config, state *state, schemaRef1, schemaRef2 *openapi3.SchemaRef, index1, index2 int) (ModifiedSubschemas, error) {

	diff, err := getSchemaDiff(config, state, schemaRef1, schemaRef2)
	if err != nil {
		return nil, err
	}
	if !diff.Empty() {
		modifiedSchemas = append(modifiedSchemas, &ModifiedSubschema{
			Base: Subschema{
				Index:     index1,
				Component: getComponentName(schemaRef1),
				Title:     schemaRef1.Value.Title,
			},
			Revision: Subschema{
				Index:     index2,
				Component: getComponentName(schemaRef2),
				Title:     schemaRef2.Value.Title,
			},
			Diff: diff,
		})
	}

	return modifiedSchemas, nil
}
