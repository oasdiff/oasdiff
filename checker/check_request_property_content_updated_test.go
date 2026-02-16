package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding contentSchema and changing contentMediaType/contentEncoding in request property
func TestRequestPropertyContentUpdated(t *testing.T) {
	s1, err := open("../data/checker/content_schema_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/content_schema_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContentUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestPropertyContentSchemaAddedId], "expected request-property-content-schema-added")
	require.True(t, ids[checker.RequestPropertyContentMediaTypeChangedId], "expected request-property-content-media-type-changed")
	require.True(t, ids[checker.RequestPropertyContentEncodingChangedId], "expected request-property-content-encoding-changed")
}
