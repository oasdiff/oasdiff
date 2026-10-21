package load

import (
	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// A refPolicy that names neither a root document nor an enforcing inner reader
// is what a new call site gets by forgetting to say which it wants. It must
// refuse, not allow: before the two cases were separate fields, a caller that
// passed no root silently got no policy at all, which is the shape of both
// GHSA-2jcc-mxv7-p3f9 and the capture-reader issue after it.
func TestRefPolicy_ZeroValueFailsClosed(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	location, err := url.Parse("http://169.254.169.254/latest/meta-data/schema.yaml")
	require.NoError(t, err)

	err = refPolicy{}.allow(loader, location)

	require.Error(t, err, "a policy that names no root must refuse every $ref")
	var extErr *ExternalRefError
	require.ErrorAs(t, err, &extErr)
}

// The inner reader owning the policy is the one case that legitimately skips
// the comparison, and it has to be chosen by name.
func TestRefPolicy_InnerEnforcesSkipsTheComparison(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	location, err := url.Parse("HEAD:openapi.yaml")
	require.NoError(t, err)

	require.NoError(t, refPolicy{innerEnforces: true}.allow(loader, location),
		"the git reader applies the policy itself")
}
