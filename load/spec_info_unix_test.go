//go:build unix

package load_test

import (
	"testing"

	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestLoadInfo_FileWindows(t *testing.T) {
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource(`C:\dev\OpenApi\spec2.yaml`))
	require.EqualError(t, err, "open C:\\dev\\OpenApi\\spec2.yaml: no such file or directory")
}

func TestLoadInfo_UriInvalid(t *testing.T) {
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("http://localhost/null"))
	require.Error(t, err)
}

func TestLoadInfo_UriBadScheme(t *testing.T) {
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("ftp://localhost/null"))
	require.EqualError(t, err, "open ftp:/localhost/null: no such file or directory")
}
