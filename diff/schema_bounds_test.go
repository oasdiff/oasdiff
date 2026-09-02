package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// boundSetters applies a value for each bound keyword, so the test can build
// present/absent schema pairs against the real getters.
var boundSetters = map[string]func(*openapi3.Schema, uint64){
	"maximum":       func(s *openapi3.Schema, v uint64) { s.Max = new(float64(v)) },
	"minimum":       func(s *openapi3.Schema, v uint64) { s.Min = new(float64(v)) },
	"multipleOf":    func(s *openapi3.Schema, v uint64) { s.MultipleOf = new(float64(v)) },
	"maxLength":     func(s *openapi3.Schema, v uint64) { s.MaxLength = &v },
	"minLength":     func(s *openapi3.Schema, v uint64) { s.MinLength = v },
	"maxItems":      func(s *openapi3.Schema, v uint64) { s.MaxItems = &v },
	"minItems":      func(s *openapi3.Schema, v uint64) { s.MinItems = v },
	"maxProperties": func(s *openapi3.Schema, v uint64) { s.MaxProps = &v },
	"minProperties": func(s *openapi3.Schema, v uint64) { s.MinProps = v },
	"minContains":   func(s *openapi3.Schema, v uint64) { s.MinContains = &v },
	"maxContains":   func(s *openapi3.Schema, v uint64) { s.MaxContains = &v },
}

func boundSchema(t *testing.T, keyword string, value uint64) *openapi3.SchemaRef {
	t.Helper()
	s := &openapi3.Schema{}
	if value != 0 {
		setter, ok := boundSetters[keyword]
		require.True(t, ok, "no setter for %s; extend boundSetters", keyword)
		setter(s, value)
	}
	return &openapi3.SchemaRef{Value: s}
}

// Each SchemaBounds row matches its getter's absence encoding: going from
// absent to a value classifies as Set, the reverse as Unset, and a change
// between two values as neither. A getter that changes how it encodes
// absence fails here.
func TestSchemaBounds(t *testing.T) {
	cfg := NewConfig()
	for _, bound := range SchemaBounds {
		absent := boundSchema(t, bound.Keyword, 0)
		low := boundSchema(t, bound.Keyword, 4)
		high := boundSchema(t, bound.Keyword, 8)

		set, err := getSchemaDiff(cfg, newState(), absent, low)
		require.NoError(t, err)
		value, ok := bound.Set(set)
		require.True(t, ok, "%s: absent to value must classify as Set", bound.Keyword)
		require.NotNil(t, value, bound.Keyword)
		_, ok = bound.Unset(set)
		require.False(t, ok, bound.Keyword)

		unset, err := getSchemaDiff(cfg, newState(), low, absent)
		require.NoError(t, err)
		value, ok = bound.Unset(unset)
		require.True(t, ok, "%s: value to absent must classify as Unset", bound.Keyword)
		require.NotNil(t, value, bound.Keyword)
		_, ok = bound.Set(unset)
		require.False(t, ok, bound.Keyword)

		changed, err := getSchemaDiff(cfg, newState(), low, high)
		require.NoError(t, err)
		_, ok = bound.Set(changed)
		require.False(t, ok, "%s: value to value is not Set", bound.Keyword)
		_, ok = bound.Unset(changed)
		require.False(t, ok, "%s: value to value is not Unset", bound.Keyword)
	}
}
