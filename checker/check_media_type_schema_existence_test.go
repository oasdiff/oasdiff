package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// A schema appearing in / disappearing from a media type that exists on both
// sides is classified by request/response contravariance (#1050):
//   - request schema added   -> ERR  (narrows accepted input)
//   - response schema removed -> ERR  (drops the output guarantee)
//   - request schema removed  -> INFO (more permissive)
//   - response schema added   -> INFO (more specific)
func TestMediaTypeSchemaExistence(t *testing.T) {
	base, err := open("../data/media-type-schema/base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/media-type-schema/revision.yaml")
	require.NoError(t, err)

	// Schema added (base -> revision): request + response media types gain a schema.
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)

	breaking := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm) // WARN+ only
	require.True(t, containsId(breaking, checker.RequestBodyMediaTypeSchemaAddedId),
		"request body schema added is breaking (narrows accepted input)")
	require.False(t, containsId(breaking, checker.ResponseBodyMediaTypeSchemaAddedId),
		"response schema added is not breaking (response got more specific)")

	all := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.True(t, containsId(all, checker.ResponseBodyMediaTypeSchemaAddedId),
		"response schema added is reported at INFO")

	// Schema removed (revision -> base): the same media types lose their schema.
	d, osm, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), revision, base)
	require.NoError(t, err)

	breaking = checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.True(t, containsId(breaking, checker.ResponseBodyMediaTypeSchemaRemovedId),
		"response schema removed is breaking (drops the output guarantee)")
	require.False(t, containsId(breaking, checker.RequestBodyMediaTypeSchemaRemovedId),
		"request schema removed is not breaking (more permissive)")

	all = checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.True(t, containsId(all, checker.RequestBodyMediaTypeSchemaRemovedId),
		"request schema removed is reported at INFO")
}

// Removing a response schema is an error, not a warning. It contains changes
// that are errors on their own: a media type with no schema constrains the body
// no more than one that never declared a property, and both
// response-media-type-removed above it and response-required-property-removed
// below it are errors, so a warning here dips in the middle of a containment
// chain. See #1140, and #1141 for the guard that would catch the class.
func TestResponseSchemaRemoved_IsAnError(t *testing.T) {
	base, err := open("../data/media-type-schema/base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/media-type-schema/revision.yaml")
	require.NoError(t, err)

	// revision -> base, so the response media types lose their schema.
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), revision, base)
	require.NoError(t, err)

	changes := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	for _, c := range changes {
		if c.GetId() == checker.ResponseBodyMediaTypeSchemaRemovedId {
			require.Equal(t, checker.ERR, c.GetLevel())
			return
		}
	}
	t.Fatal("expected a response-body-media-type-schema-removed change")
}

// a body schema arriving as the boolean `false` withdraws the media type's
// payload, so it is reported as schema-became-false on both sides rather than
// as a schema addition
func TestMediaTypeSchemaAddedAsFalse(t *testing.T) {
	s1, err := open("../data/checker/media_type_schema_degenerate_none.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/media_type_schema_degenerate_false.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.MediaTypeSchemaExistenceCheck), d, osm)

	require.Len(t, errs, 2)
	require.Equal(t, checker.ERR, requireChange(t, errs, checker.RequestBodySchemaBecameFalseId).GetLevel())
	require.Equal(t, checker.ERR, requireChange(t, errs, checker.ResponseBodySchemaBecameFalseId).GetLevel())
}

// a body schema arriving as `true` accepts what the absent schema accepted,
// so the contract is unchanged and nothing is reported; the diff still shows
// the document change
func TestMediaTypeSchemaAddedAsTrue(t *testing.T) {
	s1, err := open("../data/checker/media_type_schema_degenerate_none.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/media_type_schema_degenerate_true.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.MediaTypeSchemaExistenceCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}

// removing a `true` schema is the same non-event in the other direction
func TestMediaTypeSchemaTrueRemoved(t *testing.T) {
	s1, err := open("../data/checker/media_type_schema_degenerate_true.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/media_type_schema_degenerate_none.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.MediaTypeSchemaExistenceCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}

// removing a `false` schema leaves the media type unconstrained, which is the
// hazard the plain schema-removed rules already report
func TestMediaTypeSchemaFalseRemoved(t *testing.T) {
	s1, err := open("../data/checker/media_type_schema_degenerate_false.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/media_type_schema_degenerate_none.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.MediaTypeSchemaExistenceCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)
	require.Equal(t, checker.INFO, requireChange(t, errs, checker.RequestBodyMediaTypeSchemaRemovedId).GetLevel())
	require.Equal(t, checker.ERR, requireChange(t, errs, checker.ResponseBodyMediaTypeSchemaRemovedId).GetLevel())
}
