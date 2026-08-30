package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// a schema replaced by the boolean `false` accepts nothing: breaking on the
// request side, and for a response body too, since the client's selected
// media type can no longer be inhabited; the keyword-level echoes of the
// replacement (the type reads as removed) are suppressed in favor of the
// single transition finding per node
func TestSchemaBecameFalse(t *testing.T) {
	s1, err := open("../data/checker/schema_became_false_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/schema_became_false_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.False(t, containsId(errs, checker.RequestPropertyTypeChangedId),
		"the type removed by the false schema must not also be reported as a type change")
	require.False(t, containsId(errs, checker.ResponseBodyTypeChangedId),
		"the type removed by the false schema must not also be reported as a type change")

	require.Len(t, errs, 3)

	param := requireChange(t, errs, checker.RequestParameterSchemaBecameFalseId)
	require.Equal(t, checker.ERR, param.GetLevel())

	prop := requireChange(t, errs, checker.RequestPropertySchemaBecameFalseId)
	require.Equal(t, checker.ERR, prop.GetLevel())

	body := requireChange(t, errs, checker.ResponseBodySchemaBecameFalseId)
	require.Equal(t, checker.ERR, body.GetLevel())
}

// the reverse direction: a `false` schema replaced by one that accepts values
// widens, which is informational everywhere here: for the request side as any
// widening, and for the response body as a variant becoming usable
func TestSchemaBecameNotFalse(t *testing.T) {
	s1, err := open("../data/checker/schema_became_false_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/schema_became_false_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.Len(t, errs, 3)

	body := requireChange(t, errs, checker.ResponseBodySchemaBecameNotFalseId)
	require.Equal(t, checker.INFO, body.GetLevel())

	param := requireChange(t, errs, checker.RequestParameterSchemaBecameNotFalseId)
	require.Equal(t, checker.INFO, param.GetLevel())

	prop := requireChange(t, errs, checker.RequestPropertySchemaBecameNotFalseId)
	require.Equal(t, checker.INFO, prop.GetLevel())
}

// the transition's claims: replacing a schema that declares a type, a required
// property, an enum and a pattern with `false` produces keyword-level echoes
// with verdicts from the wrong comparison (the type reads as generalized, the
// properties as removed), which are suppressed in favor of the one transition
// finding
func TestSchemaBecameFalseClaims(t *testing.T) {
	s1, err := open("../data/checker/schema_became_false_claims_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/schema_became_false_claims_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.False(t, containsId(errs, checker.RequestPropertyRemovedId),
		"the properties on the replaced schema were not removed one by one")
	require.False(t, containsId(errs, checker.RequestBodyTypeGeneralizedId),
		"the type removed by the false schema is not a widening")

	change := requireSingleChange(t, errs, checker.RequestBodySchemaBecameFalseId)
	require.Equal(t, checker.ERR, change.GetLevel())
}

// the reverse direction: the echoes of a `false` schema being replaced read as
// narrowings (the type as changed, the required property as newly required),
// which would be spurious errors on a change that only widens
func TestSchemaBecameNotFalseClaims(t *testing.T) {
	s1, err := open("../data/checker/schema_became_false_claims_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/schema_became_false_claims_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.False(t, containsId(errs, checker.NewRequiredRequestPropertyId),
		"a property on the restored schema is not a new requirement on existing requests")
	require.False(t, containsId(errs, checker.RequestBodyTypeChangedId),
		"the type introduced by the restored schema is not a narrowing")

	change := requireSingleChange(t, errs, checker.RequestBodySchemaBecameNotFalseId)
	require.Equal(t, checker.INFO, change.GetLevel())
}

// an items schema arriving as `false` closes the tuple: arrays longer than
// prefixItems were valid and no longer are. The diff records a one-sided
// items arrival with no schema attached, so the transition is detected at the
// parent schema, at the body level and inside properties alike
func TestItemsBecameFalse(t *testing.T) {
	s1, err := open("../data/checker/items_became_false_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/items_became_false_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	req := requireChange(t, errs, checker.RequestPropertySchemaBecameFalseId)
	require.Equal(t, checker.ERR, req.GetLevel())

	resp := requireChange(t, errs, checker.ResponsePropertySchemaBecameFalseId)
	require.Equal(t, checker.INFO, resp.GetLevel())
	require.NotEmpty(t, resp.GetComment(checker.NewDefaultLocalizer()))
}

// the reverse reopens the tuple: informational on the request side as any
// widening, breaking on the response side since the server may now return
// arrays longer than the prefix
func TestItemsBecameNotFalse(t *testing.T) {
	s1, err := open("../data/checker/items_became_false_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/items_became_false_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	req := requireChange(t, errs, checker.RequestPropertySchemaBecameNotFalseId)
	require.Equal(t, checker.INFO, req.GetLevel())

	resp := requireChange(t, errs, checker.ResponsePropertySchemaBecameNotFalseId)
	require.Equal(t, checker.ERR, resp.GetLevel())
}
