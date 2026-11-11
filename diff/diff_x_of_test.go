package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

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
	require.Equal(t, diff.Subschemas{{Index: 2}}, anyOfDiff.Added)
	require.Empty(t, anyOfDiff.Deleted)
	require.Len(t, anyOfDiff.Modified, 1)
	require.Equal(t, 0, anyOfDiff.Modified[0].Base.Index)
	require.Equal(t, 0, anyOfDiff.Modified[0].Revision.Index)
	require.NotNil(t, anyOfDiff.Modified[0].Diff.PropertiesDiff.Modified["prop1"].DescriptionDiff)
}

func TestAnyOf_ExcludeDescriptions(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("anyof-base-openapi.yml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("anyof-rev-openapi.yml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(diff.WithExcludeElements([]string{diff.ExcludeDescriptionOption})), s1, s2)
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

// Annotation-only edits inside multiple untitled inline allOf branches are
// paired as modifications rather than reported as wholesale additions and
// deletions. This exercises the same pairing pass that handles anyOf and
// oneOf, applied to allOf: on main both branches would surface as 2 added +
// 2 deleted because neither branch has a title and neither is byte-identical
// across revisions.
func TestAllOf_AnnotationOnlyChanges(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("allof-annotation-only-base.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getXOfFile("allof-annotation-only-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	allOfDiff := dd.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AllOfDiff
	require.NotNil(t, allOfDiff)
	require.Empty(t, allOfDiff.Added)
	require.Empty(t, allOfDiff.Deleted)
	require.Len(t, allOfDiff.Modified, 2)

	require.Equal(t, 0, allOfDiff.Modified[0].Base.Index)
	require.Equal(t, 0, allOfDiff.Modified[0].Revision.Index)
	require.NotNil(t, allOfDiff.Modified[0].Diff.PropertiesDiff.Modified["prop1"].DescriptionDiff)

	require.Equal(t, 1, allOfDiff.Modified[1].Base.Index)
	require.Equal(t, 1, allOfDiff.Modified[1].Revision.Index)
	require.NotNil(t, allOfDiff.Modified[1].Diff.PropertiesDiff.Modified["prop2"].DescriptionDiff)
}

// Inline -> $ref refactor of an equivalent component is not reported by
// default (MatchInlineRefs=true). The Added/Deleted pair is paired up in the
// subschema matcher and removed.
func TestAnyOf_InlineRefRefactor_MatchedByDefault(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("inline-to-ref-refactor-base.yaml"))
	require.NoError(t, err)
	s2, err := loader.LoadFromFile(getXOfFile("inline-to-ref-refactor-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// The refactor reconciles away the only structural change. The whole
	// PathsDiff (or the AnyOfDiff inside it) should be absent.
	if dd.PathsDiff == nil {
		return
	}
	pathDiff, ok := dd.PathsDiff.Modified["/pets"]
	if !ok || pathDiff.OperationsDiff == nil {
		return
	}
	opDiff, ok := pathDiff.OperationsDiff.Modified["POST"]
	if !ok || opDiff.RequestBodyDiff == nil || opDiff.RequestBodyDiff.ContentDiff == nil {
		return
	}
	mediaTypeDiff, ok := opDiff.RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"]
	if !ok || mediaTypeDiff.SchemaDiff == nil {
		return
	}
	require.Nil(t, mediaTypeDiff.SchemaDiff.AnyOfDiff,
		"inline-to-$ref refactor of an equivalent component must not produce an AnyOfDiff")
}

// With --match-inline-refs=false (WithMatchInlineRefs(false)) the matcher
// falls back to the previous behaviour: the inline branch is reported as
// Deleted and the $ref branch as Added.
func TestAnyOf_InlineRefRefactor_OptOut(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("inline-to-ref-refactor-base.yaml"))
	require.NoError(t, err)
	s2, err := loader.LoadFromFile(getXOfFile("inline-to-ref-refactor-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(diff.WithMatchInlineRefs(false)), s1, s2)
	require.NoError(t, err)

	anyOfDiff := dd.PathsDiff.Modified["/pets"].OperationsDiff.Modified["POST"].RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.NotNil(t, anyOfDiff)
	require.Len(t, anyOfDiff.Added, 1, "the new $ref branch must be reported as added")
	require.Len(t, anyOfDiff.Deleted, 1, "the old inline branch must be reported as deleted")
}

// Regression: when the inline subschema on the base side and the $ref target
// on the revision side resolve byte-identical (no title or other annotation
// difference), the existing findIndenticalSchema pass used to silently
// absorb the inline side via a misplaced filter check, leaving the $ref
// alone in Added with no Deleted to pair against. The reconciliation pass
// could then not match them and the refactor was reported as a spurious
// addition. With the filter fix, the inline goes to Deleted as it should
// and the pair is reconciled. The diff is empty.
func TestAnyOf_InlineRefRefactor_ByteIdentical_MatchedByDefault(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("inline-ref-byte-identical-base.yaml"))
	require.NoError(t, err)
	s2, err := loader.LoadFromFile(getXOfFile("inline-ref-byte-identical-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	if dd.PathsDiff == nil {
		return
	}
	pathDiff, ok := dd.PathsDiff.Modified["/pets"]
	if !ok || pathDiff.OperationsDiff == nil {
		return
	}
	opDiff, ok := pathDiff.OperationsDiff.Modified["POST"]
	if !ok || opDiff.RequestBodyDiff == nil || opDiff.RequestBodyDiff.ContentDiff == nil {
		return
	}
	mediaTypeDiff, ok := opDiff.RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"]
	if !ok || mediaTypeDiff.SchemaDiff == nil {
		return
	}
	require.Nil(t, mediaTypeDiff.SchemaDiff.AnyOfDiff,
		"byte-identical inline-to-$ref refactor must be reconciled (the inline must reach Deleted so reconciliation can pair it with the Added $ref)")
}

// Pair-based matching: a standalone added $ref that happens to be equivalent
// to an existing still-present inline branch is *not* suppressed, because
// there is no deleted branch to pair it with. Verifies the matcher does not
// collapse anything that is not actually a refactor.
func TestAnyOf_StandaloneEquivalentAddedRefIsReported(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getXOfFile("inline-ref-standalone-add-base.yaml"))
	require.NoError(t, err)
	s2, err := loader.LoadFromFile(getXOfFile("inline-ref-standalone-add-revision.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	anyOfDiff := dd.PathsDiff.Modified["/pets"].OperationsDiff.Modified["POST"].RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].SchemaDiff.AnyOfDiff
	require.NotNil(t, anyOfDiff, "the standalone $ref addition is a real change and must be reported")
	require.Len(t, anyOfDiff.Added, 1, "the new $ref branch must be reported as added")
	require.Empty(t, anyOfDiff.Deleted, "nothing was removed on the base side")
}
