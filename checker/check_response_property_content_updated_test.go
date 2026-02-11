package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding contentSchema and changing contentMediaType/contentEncoding in response property
func TestResponsePropertyContentUpdated(t *testing.T) {
	s1, err := open("../data/checker/content_schema_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/content_schema_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContentUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.ResponsePropertyContentSchemaAddedId], "expected response-property-content-schema-added")
	require.True(t, ids[checker.ResponsePropertyContentMediaTypeChangedId], "expected response-property-content-media-type-changed")
	require.True(t, ids[checker.ResponsePropertyContentEncodingChangedId], "expected response-property-content-encoding-changed")
}
