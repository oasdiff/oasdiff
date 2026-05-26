package diff_test

import (
	"testing"

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
