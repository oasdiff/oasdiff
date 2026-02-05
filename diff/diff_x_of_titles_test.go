package diff_test

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func getXOfTitlesFile(file string) string {
	return fmt.Sprintf("../data/x-of-titles/%s", file)
}

func TestXOfTitles_Identical(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfTitlesFile("spec1.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfTitlesFile("spec1.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{}, s1, s2)
	require.NoError(t, err)
	require.Empty(t, dd)
}

func TestXOfTitles_TitleNameChanged(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfTitlesFile("spec1.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfTitlesFile("spec2.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{}, s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
	anyOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.ElementsMatch(t,
		diff.Subschemas{
			diff.Subschema{
				Index: 1,
				Title: "Title 3",
			},
		},
		anyOfDiff.Added)
	require.ElementsMatch(t,
		diff.Subschemas{
			diff.Subschema{
				Index: 1,
				Title: "Title 2",
			},
		},
		anyOfDiff.Deleted)
	require.Empty(t, anyOfDiff.Modified)
}

func TestXOfTitles_TitlesModified(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfTitlesFile("spec1.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfTitlesFile("spec3.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{}, s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
	anyOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.Empty(t, anyOfDiff.Added)
	require.Empty(t, anyOfDiff.Deleted)

	require.Len(t, anyOfDiff.Modified, 2)

	require.Equal(t,
		diff.Subschema{
			Index: 0,
			Title: "Title 1",
		},
		anyOfDiff.Modified[0].Base,
	)
	require.Equal(t,
		diff.Subschema{
			Index: 0,
			Title: "Title 1",
		},
		anyOfDiff.Modified[0].Revision,
	)
	require.Equal(t, []string{"string"}, anyOfDiff.Modified[0].Diff.TypeDiff.Deleted)
	require.Equal(t, []string{"boolean"}, anyOfDiff.Modified[0].Diff.TypeDiff.Added)

	require.Equal(t,
		diff.Subschema{
			Index: 1,
			Title: "Title 2",
		},
		anyOfDiff.Modified[1].Base,
	)
	require.Equal(t,
		diff.Subschema{
			Index: 1,
			Title: "Title 2",
		},
		anyOfDiff.Modified[1].Revision,
	)
	require.Equal(t, []string{"string"}, anyOfDiff.Modified[1].Diff.TypeDiff.Deleted)
	require.Equal(t, []string{"number"}, anyOfDiff.Modified[1].Diff.TypeDiff.Added)
}

func TestXOfTitles_TitlesModifiedAndAdded(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfTitlesFile("spec3.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfTitlesFile("spec4.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{}, s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
	anyOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.Len(t, anyOfDiff.Added, 1)
	require.ElementsMatch(t, diff.Subschemas{
		{
			Index: 2,
			Title: "Title 3",
		},
	}, anyOfDiff.Added)
	require.Empty(t, anyOfDiff.Deleted)
	require.Len(t, anyOfDiff.Modified, 1)
	require.Equal(t,
		diff.Subschema{
			Index: 0,
			Title: "Title 1",
		},
		anyOfDiff.Modified[0].Base,
	)
	require.Equal(t,
		diff.Subschema{
			Index: 0,
			Title: "Title 1",
		},
		anyOfDiff.Modified[0].Revision,
	)
	require.Equal(t, []string{"boolean"}, anyOfDiff.Modified[0].Diff.TypeDiff.Deleted)
	require.Equal(t, []string{"string"}, anyOfDiff.Modified[0].Diff.TypeDiff.Added)
}

func TestXOfTitles_DuplicateTitles(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfTitlesFile("spec3.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfTitlesFile("spec5.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{}, s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
	anyOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.Len(t, anyOfDiff.Added, 2)
	require.ElementsMatch(t, diff.Subschemas{
		{
			Index: 0,
			Title: "Title 2",
		},
		{
			Index: 2,
			Title: "Title 3",
		},
	}, anyOfDiff.Added)

	require.Len(t, anyOfDiff.Deleted, 1)
	require.ElementsMatch(t, diff.Subschemas{
		{
			Index: 0,
			Title: "Title 1",
		},
	}, anyOfDiff.Deleted)

	require.Empty(t, anyOfDiff.Modified, 0)
}

func TestXOfTitles_EmptyTitle(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfTitlesFile("spec6.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfTitlesFile("spec7.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{}, s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
	anyOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.Empty(t, anyOfDiff.Added)
	require.Empty(t, anyOfDiff.Deleted)
	require.Len(t, anyOfDiff.Modified, 1)
	require.Equal(t,
		diff.Subschema{
			Index: 3,
		},
		anyOfDiff.Modified[0].Base,
	)
	require.Equal(t,
		diff.Subschema{
			Index: 3,
		},
		anyOfDiff.Modified[0].Revision,
	)
	require.Equal(t, []string{"string"}, anyOfDiff.Modified[0].Diff.TypeDiff.Deleted)
	require.Equal(t, []string{"number"}, anyOfDiff.Modified[0].Diff.TypeDiff.Added)
}
