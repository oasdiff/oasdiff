package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
)

// the schema printed by "oasdiff schema" must accept the real "--format json"
// output of breaking/changelog. This guards against the schema and the output
// type drifting apart.
func TestSchema_ChangelogOutputValidatesAgainstSchema(t *testing.T) {
	// the published schema
	var schemaBuf bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff schema"), &schemaBuf, &schemaBuf))

	// a real changelog --format json output (test1 vs test3 produces many changes)
	var outBuf bytes.Buffer
	internal.Run(cmdToArgs("oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test3.yaml --format json"), &outBuf, io.Discard)
	require.NotEmpty(t, outBuf.Bytes())

	var schemaDoc any
	require.NoError(t, json.Unmarshal(schemaBuf.Bytes(), &schemaDoc))
	id, _ := schemaDoc.(map[string]any)["$id"].(string)
	require.NotEmpty(t, id)

	c := jsonschema.NewCompiler()
	require.NoError(t, c.AddResource(id, schemaDoc))
	sch, err := c.Compile(id)
	require.NoError(t, err)

	var instance any
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &instance))
	require.NoError(t, sch.Validate(instance))
}
