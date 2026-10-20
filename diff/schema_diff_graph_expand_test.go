package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// c is in a cycle with top and is reached twice on one walk, through a and
// through b. It is expanded where it is first reached and reports nothing at
// the second position, so a walk costs the size of the cycle and not the
// number of paths through it.
func TestUnroll_ExpandsACycleNodeOnce(t *testing.T) {
	top := &SchemaDiff{}
	a := &SchemaDiff{}
	b := &SchemaDiff{}
	c := &SchemaDiff{DescriptionDiff: &ValueDiff{From: "x", To: "y"}}
	top.PropertiesDiff = &SchemasDiff{Modified: ModifiedSchemasMap{"a": a, "b": b}}
	a.PropertiesDiff = &SchemasDiff{Modified: ModifiedSchemasMap{"c": c}}
	b.PropertiesDiff = &SchemasDiff{Modified: ModifiedSchemasMap{"c": c}}
	c.PropertiesDiff = &SchemasDiff{Modified: ModifiedSchemasMap{"top": top}}

	state := newState()
	for _, node := range []*SchemaDiff{top, a, b, c} {
		state.graph.nodes[valuePair{value1: &openapi3.Schema{}, value2: &openapi3.Schema{}}] = node
	}

	unrolled := newUnroller(NewConfig(), state).unroll(top)
	require.NotNil(t, unrolled.PropertiesDiff.Modified["a"].PropertiesDiff.Modified["c"],
		"the first position that reaches c reports it")
	require.Nil(t, unrolled.PropertiesDiff.Modified["b"],
		"the second position reports nothing, so b has no change below it")
}
