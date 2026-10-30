package load_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// With external refs disallowed only the root document may be read, so a $ref
// that names anything else is refused whatever alphabet or encoding it uses.
// Pinned because the comparison against the root is over paths: a normalization
// step added to it could make a location that names a different document
// compare equal to the root. A character class is not the control here either,
// as the punycode case shows: it names the same host as the case above it in
// pure ASCII.
func TestExternalRefPolicy_RefusesEncodedAndNonASCIIRefs(t *testing.T) {
	for _, tt := range []struct {
		name string
		file string
		ref  string
	}{
		{"non-ascii host", "idn.yaml", "http://例え.テスト/target.yaml#/Leak"},
		{"punycode host", "punycode.yaml", "http://xn--r8jz45g.xn--zckzah/target.yaml#/Leak"},
		{"percent-encoded traversal", "encoded.yaml", "..%2Foutside%2Ftarget.yaml#/Leak"},
		{"non-ascii file name", "nonascii.yaml", "ömlaut.yaml#/Leak"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			spec := filepath.Join(dir, tt.file)
			require.NoError(t, os.WriteFile(spec, []byte(`openapi: 3.0.0
info: {title: t, version: "1.0"}
paths:
  /a:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema: {$ref: "`+tt.ref+`"}
`), 0o644))

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = false
			_, err := load.NewSpecInfo(loader, load.NewSource(spec))
			require.Error(t, err)
			require.True(t, load.IsExternalRefDisallowed(err), "plain load: %v", err)

			captureLoader := openapi3.NewLoader()
			captureLoader.IsExternalRefsAllowed = false
			_, err = load.NewSpecInfoWithCapture(captureLoader, load.NewSource(spec))
			var extErr *load.ExternalRefError
			require.ErrorAs(t, err, &extErr, "capture reader must refuse it too")
		})
	}
}
