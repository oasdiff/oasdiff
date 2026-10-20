package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// SchemasDiff describes the changes between a pair of maps of schema objects like the components.schemas object
type SchemasDiff struct {
	Added    []string           `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string           `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedSchemasMap `json:"modified,omitempty" yaml:"modified,omitempty"`
	Base     openapi3.Schemas   `json:"-" yaml:"-"`
	Revision openapi3.Schemas   `json:"-" yaml:"-"`
}

// Empty indicates whether a change was found in this element
func (schemasDiff *SchemasDiff) Empty() bool {
	if schemasDiff == nil {
		return true
	}

	return len(schemasDiff.Added) == 0 &&
		len(schemasDiff.Deleted) == 0 &&
		len(schemasDiff.Modified) == 0
}

func newSchemasDiff() *SchemasDiff {
	return &SchemasDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedSchemasMap{},
	}
}

func getSchemasDiff(config *Config, state *state, schemas1, schemas2 openapi3.Schemas) (*SchemasDiff, error) {
	diff, err := getSchemasDiffInternal(config, state, schemas1, schemas2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getSchemasDiffInternal(config *Config, state *state, schemas1, schemas2 openapi3.Schemas) (*SchemasDiff, error) {

	result := newSchemasDiff()

	added, deleted, common := keysDiff(schemas1, schemas2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		if err := result.addModifiedSchema(config, state, name, schemas1[name], schemas2[name]); err != nil {
			return nil, err
		}
	}

	result.Base = schemas1
	result.Revision = schemas2

	return result, nil
}

func (schemasDiff *SchemasDiff) addModifiedSchema(config *Config, state *state, schemaName string, schemaRef1, schemaRef2 *openapi3.SchemaRef) error {
	return schemasDiff.Modified.addSchemaDiff(config, state, schemaName, schemaRef1, schemaRef2)
}

func (schemasDiff *SchemasDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(schemasDiff.Added),
		Deleted:  len(schemasDiff.Deleted),
		Modified: len(schemasDiff.Modified),
	}
}

// ModifiedSchemasMap is map of schema names to their respective diffs
type ModifiedSchemasMap map[string]*SchemaDiff

func (modifiedSchemas ModifiedSchemasMap) addSchemaDiff(config *Config, state *state, schemaName string, schemaRef1, schemaRef2 *openapi3.SchemaRef) error {

	diff, err := getSchemaDiff(config, state, schemaRef1, schemaRef2)
	if err != nil {
		return err
	}
	if diff != nil {
		modifiedSchemas[schemaName] = diff
	}

	return nil
}
