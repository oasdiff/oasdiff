package oasversion_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/internal/oasversion"
	"github.com/stretchr/testify/require"
)

func TestOf(t *testing.T) {
	for version, want := range map[string]oasversion.Dialect{
		"3.0.0": oasversion.OpenAPI30,
		"3.0.3": oasversion.OpenAPI30,
		"3.1.0": oasversion.OpenAPI31,
		"3.1.1": oasversion.OpenAPI31,
		"3.2.0": oasversion.OpenAPI31,
		"":      oasversion.OpenAPI30,
		"x":     oasversion.OpenAPI30,
	} {
		require.Equal(t, want, oasversion.Of(version), version)
	}
}
