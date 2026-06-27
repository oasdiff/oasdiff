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
//   - response schema removed -> WARN (drops the output guarantee)
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
