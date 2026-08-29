package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// a schema replaced by the boolean `false` accepts nothing, so on the request
// side the change is breaking, on the response side informational, and the
// keyword-level echoes of the replacement (the type reads as removed) are
// suppressed in favor of the single transition finding per node
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
	require.Equal(t, checker.INFO, body.GetLevel())
	require.NotEmpty(t, body.GetComment(checker.NewDefaultLocalizer()))
}

// the reverse direction: a `false` schema replaced by one that accepts values
// widens, which is breaking on the response side and informational on the
// request side
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
	require.Equal(t, checker.ERR, body.GetLevel())

	param := requireChange(t, errs, checker.RequestParameterSchemaBecameNotFalseId)
	require.Equal(t, checker.INFO, param.GetLevel())

	prop := requireChange(t, errs, checker.RequestPropertySchemaBecameNotFalseId)
	require.Equal(t, checker.INFO, prop.GetLevel())
}
