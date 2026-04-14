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

type schemaRefPair struct {
	SchemaRef1 *openapi3.SchemaRef
	SchemaRef2 *openapi3.SchemaRef
}

type schemaRefPairs map[string]*schemaRefPair

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

	addedSchemas, deletedSchemas, otherSchemas := diffSchemas(schemas1, schemas2)

	for schema := range addedSchemas {
		result.addAddedSchema(schema)
	}

	for schema := range deletedSchemas {
		result.addDeletedSchema(schema)
	}

	for schemaName, schemaRefPair := range otherSchemas {
		err := result.addModifiedSchema(config, state, schemaName, schemaRefPair.SchemaRef1, schemaRefPair.SchemaRef2)
		if err != nil {
			return nil, err
		}
	}

	result.Base = schemas1
	result.Revision = schemas2

	return result, nil
}

func diffSchemas(schemas1, schemas2 openapi3.Schemas) (openapi3.Schemas, openapi3.Schemas, schemaRefPairs) {

	added := openapi3.Schemas{}
	deleted := openapi3.Schemas{}
	other := schemaRefPairs{}

	for schemaName1, schemaRef1 := range schemas1 {
		schemaRef2, ok := schemas2[schemaName1]
		if !ok {
			deleted[schemaName1] = schemaRef1
			continue
		}

		other[schemaName1] = &schemaRefPair{
			SchemaRef1: schemaRef1,
			SchemaRef2: schemaRef2,
		}
	}

	for schemaName2, schemaRef2 := range schemas2 {
		_, ok := schemas1[schemaName2]
		if !ok {
			added[schemaName2] = schemaRef2
		}
	}

	return added, deleted, other
}

func (schemasDiff *SchemasDiff) addAddedSchema(schema string) {
	schemasDiff.Added = append(schemasDiff.Added, schema)
}

func (schemasDiff *SchemasDiff) addDeletedSchema(schema string) {
	schemasDiff.Deleted = append(schemasDiff.Deleted, schema)
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
	if !diff.Empty() {
		modifiedSchemas[schemaName] = diff
	}

	return nil
}
