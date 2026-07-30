package validate

import (
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/formatters"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

// A downgrade keyed by an id no rule emits would never fire, and nothing else
// would notice: RuleLevel would silently return ERR for the rule it was meant
// to cover.
func TestRuleLevels_KeysAreRegisteredRuleIDs(t *testing.T) {
	for id := range ruleLevels {
		require.True(t, slices.Contains(RuleIDs(), id),
			"ruleLevels key %q is not a registered rule id", id)
	}
}

func TestRuleLevel(t *testing.T) {
	// Explicit downgrades.
	require.Equal(t, checker.INFO, RuleLevel("example-violates-schema"))
	require.Equal(t, checker.WARN, RuleLevel("default-violates-schema"))
	require.Equal(t, checker.WARN, RuleLevel("conflicting-paths"))

	// The version-gate family is derived from the id, not listed one by one,
	// so a field gated upstream tomorrow is classified without a change here.
	require.Equal(t, checker.WARN, RuleLevel("const-field-for-3-1-plus"))
	require.Equal(t, checker.WARN, RuleLevel("webhooks-field-for-3-1-plus"))
	require.NotContains(t, ruleLevels, "const-field-for-3-1-plus")

	// Everything else is an error, including an id we've never seen: a spec
	// violation stays an error until someone decides otherwise.
	require.Equal(t, checker.ERR, RuleLevel("duplicate-operation-id"))
	require.Equal(t, checker.ERR, RuleLevel("openapi-required"))
	require.Equal(t, checker.ERR, RuleLevel("no-such-rule"))
	require.Equal(t, checker.ERR, RuleLevel(""))
}

// The severity a finding carries has to be the one `oasdiff checks validate`
// lists for its rule. These specs drive real kin errors through Validate, so
// this pins the classification end to end rather than re-asserting the map.
func TestValidate_FindingLevelsMatchTheRule(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   string
		want checker.Level
		spec string
	}{
		{
			name: "example that violates its schema is a documentation nit",
			id:   "example-violates-schema",
			want: checker.INFO,
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Age:
      type: integer
      example: "not-an-integer"
`,
		},
		{
			name: "default that violates its schema is consumed at runtime",
			id:   "default-violates-schema",
			want: checker.WARN,
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Age:
      type: integer
      default: "not-an-integer"
`,
		},
		{
			name: "a 3.1-only field in a 3.0 document is a portability problem",
			id:   "const-field-for-3-1-plus",
			want: checker.WARN,
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Fixed:
      type: string
      const: only-value
`,
		},
		{
			// oasdiff's own lints build findings directly rather than going
			// through the kin error path, so they need their own guard: the
			// listing showed "error" for these until they took their severity
			// from the same map.
			name: "an oasdiff-native lint carries the level the listing shows",
			id:   DuplicateEnumValueID,
			want: checker.WARN,
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Color:
      type: string
      enum: [red, green, red]
`,
		},
		{
			name: "a missing title is a structural break",
			id:   "info-title-required",
			want: checker.ERR,
			spec: `
openapi: 3.0.0
info: { version: "1" }
paths: {}
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			findings := Validate(mustLoad(t, tc.spec), "spec.yaml")

			idx := slices.IndexFunc(findings, func(f formatters.Finding) bool { return f.Id == tc.id })
			require.GreaterOrEqual(t, idx, 0, "expected a %s finding, got %v", tc.id, findings)
			require.Equal(t, tc.want, findings[idx].Level)
			// The listing and the finding are the same classification.
			require.Equal(t, RuleLevel(tc.id), findings[idx].Level)
		})
	}
}
