package load_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestSpecInfo_File(t *testing.T) {
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/openapi-test1.yaml"))
	require.NoError(t, err)
}

func TestLoadInfo_URI(t *testing.T) {
	data, err := os.ReadFile("../data/openapi-test1.yaml")
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data) //nolint:errcheck
	}))
	defer srv.Close()
	_, err = load.NewSpecInfo(openapi3.NewLoader(), load.NewSource(srv.URL))
	require.NoError(t, err)
}

func TestLoadInfo_Stdin(t *testing.T) {
	content := []byte(`openapi: 3.0.1
info:
  title: Test API
  version: v1
paths:
  /partner-api/test/some-method:
    get:
     responses:
       "200":
         description: Success
`)

	tmpfile, err := os.CreateTemp("", "example")
	require.NoError(t, err)

	defer os.Remove(tmpfile.Name()) //nolint:errcheck // clean up

	_, err = tmpfile.Write(content)
	require.NoError(t, err)

	_, err = tmpfile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile
	_, err = load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("-"))
	require.NoError(t, err)
}

func TestLoadInfo_NoVersion(t *testing.T) {
	content := []byte(`openapi: 3.0.1
paths:
  /partner-api/test/some-method:
    get:
     responses:
       "200":
         description: Success
`)

	tmpfile, err := os.CreateTemp("", "example")
	require.NoError(t, err)

	defer os.Remove(tmpfile.Name()) //nolint:errcheck // clean up

	_, err = tmpfile.Write(content)
	require.NoError(t, err)

	_, err = tmpfile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile
	specInfo, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("-"))
	require.NoError(t, err)
	require.Empty(t, specInfo.Version)
}

func TestSpecInfo_GlobOK(t *testing.T) {
	_, err := load.NewSpecInfoFromGlob(openapi3.NewLoader(), "../data/*.yaml")
	require.NoError(t, err)
}

func TestSpecInfo_InvalidSpec(t *testing.T) {
	_, err := load.NewSpecInfoFromGlob(openapi3.NewLoader(), "../data/ignore-err-example.txt")
	require.EqualError(t, err, "failed to load \"../data/ignore-err-example.txt\": failed to unmarshal data: json error: invalid character 'G' looking for beginning of value, yaml error: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type openapi3.TBis")
}

func TestSpecInfo_InvalidGlob(t *testing.T) {
	_, err := load.NewSpecInfoFromGlob(openapi3.NewLoader(), "[*")
	require.EqualError(t, err, "syntax error in pattern")
}

func TestSpecInfo_URL(t *testing.T) {
	_, err := load.NewSpecInfoFromGlob(openapi3.NewLoader(), "http://localhost/openapi-test1.yaml")
	require.EqualError(t, err, "no matching files (should be a glob, not a URL)")
}

func TestSpecInfo_GlobNoFiles(t *testing.T) {
	_, err := load.NewSpecInfoFromGlob(openapi3.NewLoader(), "../data/*.xxx")
	require.EqualError(t, err, "no matching files")
}

func TestSpecInfo_Options(t *testing.T) {
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/openapi-test1.yaml"), load.GetOption(load.WithFlattenAllOf(), false), load.GetOption(load.WithFlattenAllOf(), true), load.WithFlattenParams())
	require.NoError(t, err)
}

func TestSpecInfo_GlobOptions(t *testing.T) {
	_, err := load.NewSpecInfoFromGlob(openapi3.NewLoader(), "../data/*.yaml", load.WithIdentity(), load.WithFlattenAllOf(), load.WithFlattenParams())
	require.NoError(t, err)
}

// Pin the FlattenError contract: failures from WithFlattenAllOf must be
// reachable via errors.As so callers can distinguish them from genuine
// load failures (#907).
func TestSpecInfo_FlattenError_Typed(t *testing.T) {
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/allof/invalid.yaml"), load.WithFlattenAllOf())
	require.Error(t, err)

	var flatErr *load.FlattenError
	require.True(t, errors.As(err, &flatErr))
	require.Equal(t, "../data/allof/invalid.yaml", flatErr.Url)
	require.NotNil(t, flatErr.Err)

	require.EqualError(t, err, `failed to flatten allOf in "../data/allof/invalid.yaml": unable to resolve Type conflict: all Type values must be identical`)
}

func TestSpecInfo_FromData(t *testing.T) {
	data, err := os.ReadFile("../data/openapi-test1.yaml")
	require.NoError(t, err)

	specInfo, err := load.NewSpecInfoFromData(openapi3.NewLoader(), data, "myspec.yaml")
	require.NoError(t, err)
	require.NotNil(t, specInfo.Spec)
	// The provided name is the source label, not a temp path.
	require.Equal(t, "myspec.yaml", specInfo.Url)
	require.Equal(t, specInfo.Spec.Info.Version, specInfo.GetVersion())
}

func TestSpecInfo_FromData_InvalidData(t *testing.T) {
	_, err := load.NewSpecInfoFromData(openapi3.NewLoader(), []byte("not: [valid"), "myspec.yaml")
	require.Error(t, err)
}

func TestSpecInfo_FromData_Options(t *testing.T) {
	data, err := os.ReadFile("../data/openapi-test1.yaml")
	require.NoError(t, err)

	_, err = load.NewSpecInfoFromData(openapi3.NewLoader(), data, "myspec.yaml", load.WithFlattenAllOf(), load.WithFlattenParams())
	require.NoError(t, err)
}
