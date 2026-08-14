package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
)

// schemaSamples are changelog runs chosen so that between them every property
// the schema declares appears at least once. Validating output that omits a
// property proves nothing about that property, since an absent optional field
// is valid whatever the schema says about its type.
var schemaSamples = []struct {
	name string
	cmd  string
}{
	{
		"a broad change set",
		"oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test3.yaml --format json",
	},
	{
		// the only source of "attributes"
		"extensions carried into the output",
		"oasdiff changelog ../data/attributes/base.yaml ../data/attributes/revision.yaml --format json --attributes x-test",
	},
	{
		// the only source of "disclaimers"
		"a change under an unflattened allOf",
		"oasdiff changelog ../data/checker/disclaimer_allof_base.yaml ../data/checker/disclaimer_allof_revision.yaml --format json",
	},
}

// The schema printed by "oasdiff schema" must accept the real "--format json"
// output of breaking and changelog, and must describe it accurately. The schema
// is reflected from the Go types, so a type that marshals itself into something
// other than its Go shape drifts from the schema without either side failing on
// its own.
func TestSchema_ChangelogOutputValidatesAgainstSchema(t *testing.T) {
	sch := publishedSchema(t)

	covered := map[string]bool{}
	for _, sample := range schemaSamples {
		t.Run(sample.name, func(t *testing.T) {
			var out bytes.Buffer
			internal.Run(cmdToArgs(sample.cmd), &out, io.Discard)
			require.NotEmpty(t, out.Bytes())

			var instance any
			require.NoError(t, json.Unmarshal(out.Bytes(), &instance))
			require.NoError(t, sch.Validate(instance))

			for _, change := range instance.([]any) {
				for property := range change.(map[string]any) {
					covered[property] = true
				}
			}
		})
	}

	// Every declared property has now been through Validate at least once. A
	// property no sample produces needs a sample, or it is unreachable and does
	// not belong in the output type.
	for _, property := range declaredProperties(t) {
		require.Truef(t, covered[property],
			"no sample produces %q, so the schema's claim about it is never checked\n"+
				"  add a run to schemaSamples that emits it", property)
	}
}

func publishedSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	var buf bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff schema"), &buf, &buf))

	var doc any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &doc))
	id, _ := doc.(map[string]any)["$id"].(string)
	require.NotEmpty(t, id)

	c := jsonschema.NewCompiler()
	require.NoError(t, c.AddResource(id, doc))
	sch, err := c.Compile(id)
	require.NoError(t, err)
	return sch
}

func declaredProperties(t *testing.T) []string {
	t.Helper()

	var buf bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff schema"), &buf, &buf))

	var doc struct {
		Defs map[string]struct {
			Properties map[string]any `json:"properties"`
		} `json:"$defs"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &doc))

	change, ok := doc.Defs["Change"]
	require.True(t, ok, "the schema no longer defines Change")

	names := make([]string, 0, len(change.Properties))
	for name := range change.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
