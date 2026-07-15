//go:build unix

package checker_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// A change inside a schema $ref'd from another file must report that file (and
// the line inside it) as its source, not the referencing operation in the root
// document. Two external-$ref shapes are covered: a whole-file ref
// ("./user.yaml") and a ref to an arbitrary top-level key
// ("./schemas.yaml#/User", the Swagger-2-era "definitions bag").
// Unix-only: asserts on OS-native $ref path resolution.

const crossFileSourceRoot = `openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '%s'
      responses:
        "201": { description: created }
`

// crossFileTypeChange diffs two single-operation specs whose request schema is
// $ref'd from an external file, returning the one resulting change and the
// revision-side path of that external file.
func crossFileTypeChange(t *testing.T, ref, schemaFile, baseSchema, revSchema string) (checker.Change, string) {
	t.Helper()
	root := []byte(fmt.Sprintf(crossFileSourceRoot, ref))

	loadSide := func(schema string) (*load.SpecInfo, string) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, schemaFile), []byte(schema), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), root, 0644))
		loader := openapi3.NewLoader()
		loader.IncludeOrigin = true
		loader.IsExternalRefsAllowed = true
		si, err := load.NewSpecInfo(loader, load.NewSource(filepath.Join(dir, "openapi.yaml")))
		require.NoError(t, err)
		return si, filepath.Join(dir, schemaFile)
	}

	s1, _ := loadSide(baseSchema)
	s2, revSchemaPath := loadSide(revSchema)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibility(checker.NewConfig(checker.GetAllChecks()), d, osm)
	require.Len(t, changes, 1)
	require.Equal(t, "request-property-type-changed", changes[0].GetId())
	return changes[0], revSchemaPath
}

// Whole-file $ref: works today; pins the correct behavior.
func TestCrossFileSource_WholeFileRef(t *testing.T) {
	change, revSchemaPath := crossFileTypeChange(t, "./user.yaml", "user.yaml",
		"type: object\nproperties:\n  id:\n    type: string\n",
		"type: object\nproperties:\n  id:\n    type: integer\n")

	src := change.GetRevisionSource()
	require.NotNil(t, src)
	require.Equal(t, revSchemaPath, src.File,
		"the source must be the external file the property lives in")
	require.Equal(t, 4, src.Line, "the changed type line inside user.yaml")
}

// Arbitrary top-level key $ref: the schema lives under "User" in schemas.yaml,
// so the change's source must be schemas.yaml, not the referencing operation in
// the root document.
func TestCrossFileSource_ArbitraryTopLevelKeyRef(t *testing.T) {
	change, revSchemaPath := crossFileTypeChange(t, "./schemas.yaml#/User", "schemas.yaml",
		"User:\n  type: object\n  properties:\n    id:\n      type: string\n",
		"User:\n  type: object\n  properties:\n    id:\n      type: integer\n")

	src := change.GetRevisionSource()
	require.NotNil(t, src)
	require.Equal(t, revSchemaPath, src.File,
		"the source must be the external file the property lives in, not the root document")
	require.Equal(t, 5, src.Line, "the changed type line inside schemas.yaml")
}
