//go:build unix

package load_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestLoadInfo_FileWindows(t *testing.T) {
	_, err := load.NewSpecInfo(MockLoader{}, load.NewSource(`C:\dev\OpenApi\spec2.yaml`))
	require.EqualError(t, err, "open C:\\dev\\OpenApi\\spec2.yaml: no such file or directory")
}

func TestLoadInfo_UriInvalid(t *testing.T) {
	_, err := load.NewSpecInfo(MockLoader{}, load.NewSource("http://localhost/null"))
	require.EqualError(t, err, "open ../null: no such file or directory")
}

func TestLoadInfo_UriBadScheme(t *testing.T) {
	_, err := load.NewSpecInfo(MockLoader{}, load.NewSource("ftp://localhost/null"))
	require.EqualError(t, err, "open ftp:/localhost/null: no such file or directory")
}
