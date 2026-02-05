package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestOpenAPI31_JSONSchemaDialect(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.JSONSchemaDialectDiff)
	require.Equal(t, "https://json-schema.org/draft/2020-12/schema", d.JSONSchemaDialectDiff.From)
	require.Equal(t, "https://json-schema.org/draft/2020-12/schema#", d.JSONSchemaDialectDiff.To)
}

func TestOpenAPI31_InfoSummary(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.InfoDiff)
	require.NotNil(t, d.InfoDiff.SummaryDiff)
	require.Equal(t, "Base API for testing OpenAPI 3.1 features", d.InfoDiff.SummaryDiff.From)
	require.Equal(t, "Revised API for testing OpenAPI 3.1 features", d.InfoDiff.SummaryDiff.To)
}

func TestOpenAPI31_LicenseIdentifier(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.InfoDiff)
	require.NotNil(t, d.InfoDiff.LicenseDiff)
	require.NotNil(t, d.InfoDiff.LicenseDiff.IdentifierDiff)
	require.Equal(t, "MIT", d.InfoDiff.LicenseDiff.IdentifierDiff.From)
	require.Equal(t, "Apache-2.0", d.InfoDiff.LicenseDiff.IdentifierDiff.To)
}

func TestOpenAPI31_SchemaConst(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.ComponentsDiff)
	require.NotNil(t, d.ComponentsDiff.SchemasDiff)
	require.NotNil(t, d.ComponentsDiff.SchemasDiff.Modified)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)
	require.NotNil(t, testSchemaDiff.PropertiesDiff)
	require.NotNil(t, testSchemaDiff.PropertiesDiff.Modified)

	statusDiff := testSchemaDiff.PropertiesDiff.Modified["status"]
	require.NotNil(t, statusDiff)
	require.NotNil(t, statusDiff.ConstDiff)
	require.Equal(t, "active", statusDiff.ConstDiff.From)
	require.Equal(t, "inactive", statusDiff.ConstDiff.To)
}

func TestOpenAPI31_SchemaExamples(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.ComponentsDiff)
	require.NotNil(t, d.ComponentsDiff.SchemasDiff)
	require.NotNil(t, d.ComponentsDiff.SchemasDiff.Modified)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)
	require.NotNil(t, testSchemaDiff.ExamplesDiff)
}

func TestOpenAPI31_SchemaPrefixItems(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.ComponentsDiff)
	require.NotNil(t, d.ComponentsDiff.SchemasDiff)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)
	require.NotNil(t, testSchemaDiff.PropertiesDiff)
	require.NotNil(t, testSchemaDiff.PropertiesDiff.Modified)

	tagsDiff := testSchemaDiff.PropertiesDiff.Modified["tags"]
	require.NotNil(t, tagsDiff)
	require.NotNil(t, tagsDiff.PrefixItemsDiff)
}

func TestOpenAPI31_SchemaContains(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)

	tagsDiff := testSchemaDiff.PropertiesDiff.Modified["tags"]
	require.NotNil(t, tagsDiff)
	require.NotNil(t, tagsDiff.ContainsDiff)
	require.NotNil(t, tagsDiff.MinContainsDiff)
	require.NotNil(t, tagsDiff.MaxContainsDiff)
}

func TestOpenAPI31_SchemaPatternProperties(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)

	metadataDiff := testSchemaDiff.PropertiesDiff.Modified["metadata"]
	require.NotNil(t, metadataDiff)
	require.NotNil(t, metadataDiff.PatternPropertiesDiff)
}

func TestOpenAPI31_SchemaDependentSchemas(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)

	metadataDiff := testSchemaDiff.PropertiesDiff.Modified["metadata"]
	require.NotNil(t, metadataDiff)
	require.NotNil(t, metadataDiff.DependentSchemasDiff)
}

func TestOpenAPI31_SchemaPropertyNames(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)

	metadataDiff := testSchemaDiff.PropertiesDiff.Modified["metadata"]
	require.NotNil(t, metadataDiff)
	require.NotNil(t, metadataDiff.PropertyNamesDiff)
}

func TestOpenAPI31_SchemaUnevaluatedProperties(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)

	metadataDiff := testSchemaDiff.PropertiesDiff.Modified["metadata"]
	require.NotNil(t, metadataDiff)
	require.NotNil(t, metadataDiff.UnevaluatedPropertiesDiff)
}

func TestOpenAPI31_SchemaUnevaluatedItems(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)

	testSchemaDiff := d.ComponentsDiff.SchemasDiff.Modified["TestSchema"]
	require.NotNil(t, testSchemaDiff)

	itemsDiff := testSchemaDiff.PropertiesDiff.Modified["items"]
	require.NotNil(t, itemsDiff)
	require.NotNil(t, itemsDiff.UnevaluatedItemsDiff)
}
