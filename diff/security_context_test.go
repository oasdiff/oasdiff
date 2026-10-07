package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func loadSecurityFixture(t *testing.T, name string) *openapi3.T {
	t.Helper()
	spec, err := openapi3.NewLoader().LoadFromFile("../data/checker/api_security_anonymous_" + name + ".yaml")
	require.NoError(t, err)
	return spec
}

func TestSecurityContext(t *testing.T) {
	base := loadSecurityFixture(t, "none")
	revision := loadSecurityFixture(t, "global_api_key")

	d, err := diff.Get(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	require.Empty(t, d.SecurityContext.Base)
	require.Equal(t, revision.Security, d.SecurityContext.Revision)
	require.Equal(t, []diff.OperationPair{
		{Path: "/health", Method: "GET", Base: base.Paths.Find("/health").Get, Revision: revision.Paths.Find("/health").Get},
		{Path: "/pets", Method: "GET", Base: base.Paths.Find("/pets").Get, Revision: revision.Paths.Find("/pets").Get},
	}, d.SecurityContext.Operations)
}
