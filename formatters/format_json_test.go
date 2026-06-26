package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

var jsonFormatter = formatters.JSONFormatter{
	Localizer: MockLocalizer,
}

func TestJsonLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatJSON), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.JSONFormatter{}, f)
}

func TestJsonFormatter_RenderChangelog(t *testing.T) {
	testChanges := checker.Changes{
		checker.ComponentChange{
			Id:    "change_id",
			Level: checker.ERR,
		},
	}

	out, err := jsonFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	require.NoError(t, err)
	require.Equal(t, "[{\"id\":\"change_id\",\"text\":\"This is a breaking change.\",\"level\":3,\"section\":\"components\",\"fingerprint\":\"f8cd9af6b117\"}]", string(out))
}

func TestJsonFormatter_RenderChecks(t *testing.T) {
	checks := formatters.Checks{
		{
			Id:          "change_id",
			Level:       "info",
			Direction:   "request",
			Area:        "schema",
			Kind:        "existence",
			Action:      "remove",
			Description: "This is a breaking change.",
			Mitigation:  "Fix it.",
		},
	}

	out, err := jsonFormatter.RenderChecks(checks, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, `[{"id":"change_id","level":"info","direction":"request","area":"schema","kind":"existence","action":"remove","description":"This is a breaking change.","mitigation":"Fix it."}]`, string(out))
}

// An empty (nil) diff renders as the empty JSON object, not empty bytes,
// so the output is always valid JSON.
func TestJsonFormatter_RenderDiff(t *testing.T) {
	out, err := jsonFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, "{}", string(out))
}

// A nil pointer-shaped value (here a nil spec) renders as the empty object,
// not empty bytes, so the output is valid JSON. Note this is "{}" rather than
// a marshaled zero value: a zeroed openapi3.T would marshal to
// {"info":null,"openapi":"","paths":null}.
func TestJsonFormatter_RenderFlatten(t *testing.T) {
	out, err := jsonFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, "{}", string(out))
}

// A nil slice-shaped value renders as the empty array, not empty bytes.
func TestJsonFormatter_RenderValidate_Nil(t *testing.T) {
	out, err := jsonFormatter.RenderValidate(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, "[]", string(out))
}

func TestJsonFormatter_RenderSummary(t *testing.T) {
	out, err := jsonFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, `{"diff":false}`, string(out))
}

func TestJsonFormatter_RenderChangelog_WrappedCarriesDiffEmpty(t *testing.T) {
	// When WrapInObject is set, the JSON wrapper carries diff_empty alongside
	// the changes list so HTTP consumers can distinguish "specs are identical"
	// from "specs differ but no rule fired."
	emptyOut, err := jsonFormatter.RenderChangelog(checker.Changes{}, formatters.RenderOpts{WrapInObject: true, DiffEmpty: true}, "", "")
	require.NoError(t, err)
	require.JSONEq(t, `{"changes":[], "diff_empty":true}`, string(emptyOut))

	differOut, err := jsonFormatter.RenderChangelog(checker.Changes{}, formatters.RenderOpts{WrapInObject: true, DiffEmpty: false}, "", "")
	require.NoError(t, err)
	require.JSONEq(t, `{"changes":[], "diff_empty":false}`, string(differOut))
}

func TestJsonFormatter_RenderChangelog_BareArrayIgnoresOpts(t *testing.T) {
	// Without WrapInObject the output is a bare JSON array. DiffEmpty has
	// nowhere to live and is ignored, preserving the existing CLI shape.
	out, err := jsonFormatter.RenderChangelog(checker.Changes{}, formatters.RenderOpts{DiffEmpty: false}, "", "")
	require.NoError(t, err)
	require.Equal(t, "[]", string(out))
}

func TestJsonFormatter_RenderChangelog_WithSources(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Id:        "change_id", // Use ID that MockLocalizer recognizes
			Level:     checker.ERR,
			Operation: "GET",
			Path:      "/test",
			Source:    &load.Source{}, // Need this to avoid nil pointer dereference
			CommonChange: checker.CommonChange{
				BaseSource:     checker.NewSource("base.yaml", 10, 5),
				RevisionSource: checker.NewEmptySource(), // Empty - API was removed
			},
		},
	}

	out, err := jsonFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	require.NoError(t, err)

	// Parse and check that baseSource is included but revisionSource is not (due to omitempty)
	expected := `[{"id":"change_id","text":"This is a breaking change.","level":3,"operation":"GET","path":"/test","section":"paths","baseSource":{"file":"base.yaml","line":10,"column":5},"fingerprint":"80a8c624dc40"}]`
	require.Equal(t, expected, string(out))
}
