package diff_test

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func getXOfFile(file string) string {
	return fmt.Sprintf("../data/x-of/%s", file)
}

func TestAllOf_SingleRef(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("single-ref-base.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("single-ref-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	allOfDiff := dd.PathsDiff.Modified["/api"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AllOfDiff
	require.Len(t, allOfDiff.Modified, 1)
	require.Equal(t, diff.Subschema{
		Index:     0,
		Component: "ProductDto",
	}, allOfDiff.Modified[0].Base)
	require.Equal(t, diff.Subschema{
		Index:     0,
		Component: "ProductDto",
	}, allOfDiff.Modified[0].Revision)
	require.Equal(t, []string{"sku"}, allOfDiff.Modified[0].Diff.PropertiesDiff.Added)
}

func TestOneOf_TwoRefs(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("two-refs-base.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("two-refs-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	oneOfDiff := dd.PathsDiff.Modified["/pets"].OperationsDiff.Modified["PATCH"].RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].SchemaDiff.OneOfDiff
	require.Len(t, oneOfDiff.Modified, 1)
	require.Equal(t, diff.Subschema{
		Index:     1,
		Component: "Dog",
	}, oneOfDiff.Modified[0].Base)
	require.Equal(t, diff.Subschema{
		Index:     1,
		Component: "Dog",
	}, oneOfDiff.Modified[0].Revision)
	require.Equal(t, 1, oneOfDiff.Modified[0].Diff.AllOfDiff.Modified[0].Base.Index)
	require.Equal(t, []string{"guard"}, oneOfDiff.Modified[0].Diff.AllOfDiff.Modified[0].Diff.PropertiesDiff.Added)
}

func TestOneOf_ChangeBoth(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("two-refs-base.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("two-refs-both-changed-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	oneOfDiff := dd.PathsDiff.Modified["/pets"].OperationsDiff.Modified["PATCH"].RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].SchemaDiff.OneOfDiff
	require.Len(t, oneOfDiff.Modified, 2)

	require.Equal(t, diff.Subschema{
		Index:     0,
		Component: "Cat",
	}, oneOfDiff.Modified[0].Base)
	require.Equal(t, diff.Subschema{
		Index:     0,
		Component: "Cat",
	}, oneOfDiff.Modified[0].Revision)
	require.Equal(t, 1, oneOfDiff.Modified[0].Diff.AllOfDiff.Modified[0].Base.Index)
	require.Equal(t, []string{"miao"}, oneOfDiff.Modified[0].Diff.AllOfDiff.Modified[0].Diff.PropertiesDiff.Added)

	require.Equal(t, diff.Subschema{
		Index:     1,
		Component: "Dog",
	}, oneOfDiff.Modified[1].Base)
	require.Equal(t, diff.Subschema{
		Index:     1,
		Component: "Dog",
	}, oneOfDiff.Modified[1].Revision)
	require.Equal(t, 1, oneOfDiff.Modified[1].Diff.AllOfDiff.Modified[0].Base.Index)
	require.Equal(t, []string{"guard"}, oneOfDiff.Modified[1].Diff.AllOfDiff.Modified[0].Diff.PropertiesDiff.Added)
}

func TestOneOf_TwoInlineDuplicate(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("two-inline-base.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("two-inline-revision-duplicate.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	oneOfDiff := dd.PathsDiff.Modified["/api"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.OneOfDiff
	require.Len(t, oneOfDiff.Modified, 1)
	require.Equal(t, 0, oneOfDiff.Modified[0].Base.Index)
	require.Equal(t, 1, oneOfDiff.Modified[0].Revision.Index)
	require.Equal(t, "name2", oneOfDiff.Modified[0].Diff.PropertiesDiff.Added[0])
	require.Equal(t, "name1", oneOfDiff.Modified[0].Diff.PropertiesDiff.Deleted[0])
}

func TestOneOf_TwoInlineOneModified(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("two-inline-base.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("two-inline-revision-one-modified.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	oneOfDiff := dd.PathsDiff.Modified["/api"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.OneOfDiff
	require.Len(t, oneOfDiff.Modified, 1)
	require.Equal(t, 0, oneOfDiff.Modified[0].Base.Index)
	require.Equal(t, 1, oneOfDiff.Modified[0].Revision.Index)
	require.Equal(t, "name4", oneOfDiff.Modified[0].Diff.PropertiesDiff.Added[0])
	require.Equal(t, "name1", oneOfDiff.Modified[0].Diff.PropertiesDiff.Deleted[0])
}

func TestOneOf_MultiRefs(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("multi-refs-base.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("multi-refs-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	oneOfDiff := dd.PathsDiff.Modified["/pets"].OperationsDiff.Modified["GET"].RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].SchemaDiff.OneOfDiff
	require.Len(t, oneOfDiff.Modified, 1)
	require.Equal(t, diff.Subschema{
		Index:     2,
		Component: "Dog",
	}, oneOfDiff.Modified[0].Base)
	require.Equal(t, diff.Subschema{
		Index:     1,
		Component: "Dog",
	}, oneOfDiff.Modified[0].Revision)
	require.Equal(t, "bark", oneOfDiff.Modified[0].Diff.PropertiesDiff.Added[0])
	require.Equal(t, "name", oneOfDiff.Modified[0].Diff.PropertiesDiff.Deleted[0])
}

func TestAnyOf_IncludeDescriptions(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("anyof-base-openapi.yml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("anyof-rev-openapi.yml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	anyOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.ElementsMatch(t, diff.Subschemas{
		{
			Index: 0,
		},
		{
			Index: 2,
		},
	}, anyOfDiff.Added)
	require.ElementsMatch(t, diff.Subschemas{
		{
			Index: 0,
		},
	}, anyOfDiff.Deleted)
	require.Empty(t, anyOfDiff.Modified)
}

func TestAnyOf_ExcludeDescriptions(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("anyof-base-openapi.yml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("anyof-rev-openapi.yml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig().WithExcludeElements([]string{diff.ExcludeDescriptionOption}), s1, s2)
	require.NoError(t, err)
	anyOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.ElementsMatch(t, diff.Subschemas{
		{
			Index: 2,
		},
	}, anyOfDiff.Added)
	require.Empty(t, anyOfDiff.Deleted)
	require.Empty(t, anyOfDiff.Modified)
}
