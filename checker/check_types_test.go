package checker

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func breaking(t *testing.T, typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, isJson bool, revisionTypes *openapi3.Types) {
	t.Helper()

	mediaType := ""
	if isJson {
		mediaType = "application/json"
	}

	require.True(t, typeOrFormatBreaking(typeDiff, formatDiff, isStronglyTyped(mediaType), revisionTypes), "typeOrFormatBreaking failed")
}

func notBreaking(t *testing.T, typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, isJson bool, revisionTypes *openapi3.Types) {
	t.Helper()

	mediaType := ""
	if isJson {
		mediaType = "application/json"
	}
	require.False(t, typeOrFormatBreaking(typeDiff, formatDiff, isStronglyTyped(mediaType), revisionTypes), "typeOrFormatBreaking failed")
}

func TestStringtoInt(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"string"},
		Added:   []string{"integer"},
	}

	var formatDiff *diff.ValueDiff

	revisionType := &openapi3.Types{
		"integer",
	}

	breaking(t, typeDiff, formatDiff, false, revisionType)
}

func TestIntToString(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"integer"},
		Added:   []string{"string"},
	}

	var formatDiff *diff.ValueDiff

	revisionType := &openapi3.Types{
		"string",
	}
	notBreaking(t, typeDiff, formatDiff, false, revisionType)
}

func TestTypeDeleted(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"integer"},
		Added:   nil,
	}

	var formatDiff *diff.ValueDiff

	revisionType := &openapi3.Types{}

	notBreaking(t, typeDiff, formatDiff, false, revisionType)
}

func TestIntToStringJson(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"integer"},
		Added:   []string{"string"},
	}

	var formatDiff *diff.ValueDiff

	revisionType := &openapi3.Types{
		"string",
	}
	breaking(t, typeDiff, formatDiff, true, revisionType)
}

func TestIntToNumber(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"integer"},
		Added:   []string{"number"},
	}

	var formatDiff *diff.ValueDiff

	revisionType := &openapi3.Types{
		"number",
	}
	notBreaking(t, typeDiff, formatDiff, false, revisionType)
}

func TestUnchanged(t *testing.T) {
	var typeDiff *diff.StringsDiff
	var formatDiff *diff.ValueDiff

	revisionType := &openapi3.Types{
		"integer",
	}
	notBreaking(t, typeDiff, formatDiff, false, revisionType)
}

func TestFormatAdded(t *testing.T) {
	var typeDiff *diff.StringsDiff
	var formatDiff = &diff.ValueDiff{
		From: nil,
		To:   "int64",
	}

	revisionType := &openapi3.Types{
		"string",
	}
	breaking(t, typeDiff, formatDiff, false, revisionType)
}

// The type axis and the format axis are breaking independently: a breaking
// format change makes the overall change breaking even when the type change on
// its own is not. Here the parameter type changes integer -> string, which is
// not breaking on its own (the value is a string on the wire either way), but
// the added date-time format rejects values that were previously accepted (e.g.
// "123"), so the overall change is breaking.
func TestBreakingFormatNotMaskedByTypeChange(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"integer"},
		Added:   []string{"string"},
	}
	formatDiff := &diff.ValueDiff{
		From: nil,
		To:   "date-time",
	}
	revisionType := &openapi3.Types{
		"string",
	}
	breaking(t, typeDiff, formatDiff, false, revisionType)
}

// Removing a format constraint is a non-breaking generalization for any revision
// type, and it stays non-breaking when the type also changes. Here the type
// generalizes (integer -> number) and the format is removed (int64 -> ""), so
// the overall change is non-breaking.
func TestFormatRemovedWithTypeChangeNotBreaking(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"integer"},
		Added:   []string{"number"},
	}
	formatDiff := &diff.ValueDiff{
		From: "int64",
		To:   "",
	}
	revisionType := &openapi3.Types{
		"number",
	}
	notBreaking(t, typeDiff, formatDiff, false, revisionType)
}

// Removing the type constraint from a request body is a non-breaking
// generalization: the server then accepts any value. This holds even for a
// strongly-typed (JSON) media type, where isTypeContained alone would not grant
// it.
func TestRequestTypeConstraintRemovedNotBreaking(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Deleted: []string{"object"},
	}
	schemaDiff := &diff.SchemaDiff{
		Base:     &openapi3.Schema{Type: &openapi3.Types{"object"}},
		Revision: &openapi3.Schema{},
		TypeDiff: typeDiff,
	}
	require.False(t, requestTypeFormatBreaking(typeDiff, nil, "application/json", schemaDiff))
}

// Adding a type constraint to a response body where there was none is a
// non-breaking specialization: the server then returns a subset of what it
// could before. This is the covariant mirror of removing the constraint on the
// request side, and likewise holds for a strongly-typed (JSON) media type.
func TestResponseTypeConstraintAddedNotBreaking(t *testing.T) {
	typeDiff := &diff.StringsDiff{
		Added: []string{"object"},
	}
	schemaDiff := &diff.SchemaDiff{
		Base:     &openapi3.Schema{},
		Revision: &openapi3.Schema{Type: &openapi3.Types{"object"}},
		TypeDiff: typeDiff,
	}
	require.False(t, responseTypeFormatBreaking(typeDiff, nil, "application/json", schemaDiff))
}

func TestIsJsonMediaType(t *testing.T) {
	require.True(t, isJsonMediaType("application/json"))
	require.True(t, isJsonMediaType("application/problem+json"))
	require.True(t, isJsonMediaType("application/vnd.api+json"))
	require.True(t, isJsonMediaType("application/any-string+json"))
	require.False(t, isJsonMediaType("application/xml"))
	require.False(t, isJsonMediaType("text/plain"))
	require.False(t, isJsonMediaType("application/json-patch")) // Note: Differs from application/json-patch+json
	require.False(t, isJsonMediaType(""))
}
