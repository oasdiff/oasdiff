package diff_test

import (
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestSchemaRefsValidationEquivalent_IgnoresTitle(t *testing.T) {
	base := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"string"},
			Enum: []any{"user", "superadmin"},
		},
	}
	revision := &openapi3.SchemaRef{
		Ref: "#/components/schemas/UserRole",
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{"string"},
			Enum:        []any{"user", "superadmin"},
			Title:       "UserRole",
			Description: "Named role enum",
			Default:     "user",
			Example:     "superadmin",
			Examples:    []any{"user", "superadmin"},
			Comment:     "generated from a named enum",
		},
	}

	require.True(t, diff.SchemaRefsValidationEquivalent(diff.NewConfig(), base, revision))
}

func TestSchemaRefsValidationEquivalent_DetectsValidationChange(t *testing.T) {
	base := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"string"},
			Enum: []any{"user", "superadmin"},
		},
	}
	revision := &openapi3.SchemaRef{
		Ref: "#/components/schemas/UserRole",
		Value: &openapi3.Schema{
			Type:  &openapi3.Types{"string"},
			Enum:  []any{"user"},
			Title: "UserRole",
		},
	}

	require.False(t, diff.SchemaRefsValidationEquivalent(diff.NewConfig(), base, revision))
}

func TestSchemaRefsValidationEquivalent_DetectsDeprecatedChange(t *testing.T) {
	base := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"role": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
						Enum: []any{"user", "superadmin"},
					},
				},
			},
		},
	}
	revision := &openapi3.SchemaRef{
		Ref: "#/components/schemas/UserPayload",
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"role": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:       &openapi3.Types{"string"},
						Enum:       []any{"user", "superadmin"},
						Deprecated: true,
					},
				},
			},
		},
	}

	require.False(t, diff.SchemaRefsValidationEquivalent(diff.NewConfig(), base, revision))
}

func TestSchemaRefsValidationEquivalent_NilRefs(t *testing.T) {
	schema := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}

	require.True(t, diff.SchemaRefsValidationEquivalent(diff.NewConfig(), nil, nil))
	require.False(t, diff.SchemaRefsValidationEquivalent(diff.NewConfig(), nil, schema))
	require.False(t, diff.SchemaRefsValidationEquivalent(diff.NewConfig(), schema, nil))
}

// Node.child is a $ref to Link in the base and Link's body inlined in the
// revision, and Link.target points back at Node. Matching the two branches
// compares them for equivalence, which is a schema diff of its own that
// reaches Node.child again and asks the same question. The nested comparison
// must decline rather than start another; declining reads as a difference,
// so the refactor is reported instead of reconciled, and the diff finishes.
func TestSchemaRefsValidationEquivalent_CyclicInlineRefRefactorTerminates(t *testing.T) {
	loader := openapi3.NewLoader()
	base, err := loader.LoadFromFile("../data/circular-inline-ref1.yaml")
	require.NoError(t, err)
	revision, err := loader.LoadFromFile("../data/circular-inline-ref2.yaml")
	require.NoError(t, err)

	type result struct {
		d   *diff.Diff
		err error
	}
	done := make(chan result, 1)
	go func() {
		d, err := diff.Get(diff.NewConfig(), base, revision)
		done <- result{d, err}
	}()

	select {
	case r := <-done:
		require.NoError(t, r.err)
		require.Contains(t, r.d.ComponentsDiff.SchemasDiff.Modified, "Node")
	case <-time.After(30 * time.Second):
		t.Fatal("the diff of the cyclic inline-ref refactor did not finish within 30s")
	}
}
