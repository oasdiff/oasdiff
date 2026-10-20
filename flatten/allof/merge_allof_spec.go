package allof

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// MergeSpec merges all instances of allOf in place, across every schema in
// the document, and gives every cycle the merge anchors a $ref
// (nameAnchoredCycles), so the merged document always serializes. One merge
// state serves the whole document, so a schema reached from several
// attachment points is merged once and stays one object: every $ref to it
// keeps sharing one Value, as in the input.
func MergeSpec(spec *openapi3.T) (*openapi3.T, error) {
	state := newState()
	err := spec.WalkSchemas(func(_ string, s *openapi3.SchemaRef) error {
		if _, err := mergeInternal(state, s); err != nil {
			return err
		}
		return openapi3.SkipSubtree
	})
	if err != nil {
		return spec, err
	}
	if err := mergeCircular(state); err != nil {
		return spec, err
	}
	writeBack(spec, state)
	nameAnchoredCycles(spec, state.anchored, state.hints)
	return spec, nil
}

// writeBack writes each merge into the object it was merged from and points
// every reference to a merged copy back at that object, so the merged
// document reuses the input's objects: a copy referenced from many places
// would otherwise multiply into one copy per attachment point.
func writeBack(spec *openapi3.T, state *state) {
	originals := map[*openapi3.Schema]*openapi3.Schema{}
	for original, merged := range state.mergedSchemas {
		// the merge may have replaced its own result value (anchorInFlight);
		// the last replacement is the merge, and every value along the way
		// is referenced somewhere
		for {
			originals[merged] = original
			replacement, ok := state.replaced[merged]
			if !ok {
				break
			}
			merged = replacement
		}
		*original = *merged
		if hint, ok := state.hints[merged]; ok {
			state.hints[original] = hint
		}
	}
	redirectSchemaRefs(reflect.ValueOf(spec), originals, map[*openapi3.Schema]bool{})
}

// nameAnchoredCycles gives every anchored back-edge a $ref. The edge's value
// is right (a cycle in the input is a cycle in the merged output), but the
// edge carries no $ref, and a ref-less cycle has no serialized form. A target
// that is a named component gets that name; an anonymous target is hoisted
// into components.schemas under a name built from the component names it
// merges (AllOfMerged_NodeA_NodeB), so the flattened output is identical
// whether flatten runs standalone or inside diff, and the same logical cycle
// keeps its name when unrelated parts of the document change. Collisions and
// hintless targets fall back to a numeric suffix, assigned in the document's
// walk order, which is deterministic.
func nameAnchoredCycles(spec *openapi3.T, anchored []*openapi3.SchemaRef, hints map[*openapi3.Schema]string) {
	if len(anchored) == 0 {
		return
	}

	nameByValue := map[*openapi3.Schema]string{}
	if spec.Components != nil {
		for name, ref := range spec.Components.Schemas {
			if ref != nil && ref.Value != nil {
				nameByValue[ref.Value] = name
			}
		}
	}

	walkOrder := map[*openapi3.Schema]int{}
	_ = spec.WalkSchemas(func(_ string, s *openapi3.SchemaRef) error {
		walkOrder[s.Value] = len(walkOrder)
		return nil
	})
	slices.SortStableFunc(anchored, func(a, b *openapi3.SchemaRef) int {
		return walkOrder[a.Value] - walkOrder[b.Value]
	})

	namer := componentNamer{spec: spec, next: 1}
	for _, edge := range anchored {
		name, ok := nameByValue[edge.Value]
		if !ok {
			name = namer.name(hints[edge.Value])
			if spec.Components == nil {
				spec.Components = &openapi3.Components{}
			}
			if spec.Components.Schemas == nil {
				spec.Components.Schemas = openapi3.Schemas{}
			}
			spec.Components.Schemas[name] = openapi3.NewSchemaRef("", edge.Value)
			nameByValue[edge.Value] = name
		}
		edge.Ref = "#/components/schemas/" + name
	}
}

// componentNamer builds names for hoisted cycle components that are free in
// the spec's components section.
type componentNamer struct {
	spec *openapi3.T
	next int
}

// name returns AllOfMerged_<hint>, numerically suffixed past collisions, or
// the next free numeric AllOfMergedN when there is no hint. Hint-derived
// names depend only on what was merged, so they are stable across revisions;
// the numeric forms depend on the caller's naming order.
func (n *componentNamer) name(hint string) string {
	if hint != "" {
		name := "AllOfMerged_" + hint
		for suffix := 2; n.taken(name); suffix++ {
			name = fmt.Sprintf("AllOfMerged_%s_%d", hint, suffix)
		}
		return name
	}
	for {
		name := fmt.Sprintf("AllOfMerged%d", n.next)
		n.next++
		if !n.taken(name) {
			return name
		}
	}
}

func (n *componentNamer) taken(name string) bool {
	return n.spec.Components != nil && n.spec.Components.Schemas[name] != nil
}

// redirectSchemaRefs walks every SchemaRef reachable from v and points those
// whose Value is a key of to at the key's value instead. The traversal is
// type-driven so a new schema field carrying subschemas is covered without
// being listed here.
func redirectSchemaRefs(v reflect.Value, to map[*openapi3.Schema]*openapi3.Schema, seen map[*openapi3.Schema]bool) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		if ref, ok := v.Interface().(*openapi3.SchemaRef); ok {
			if target, ok := to[ref.Value]; ok {
				ref.Value = target
			}
			redirectSchemaRefs(reflect.ValueOf(ref.Value), to, seen)
			return
		}
		if schema, ok := v.Interface().(*openapi3.Schema); ok {
			if schema == nil || seen[schema] {
				return
			}
			seen[schema] = true
		}
		redirectSchemaRefs(v.Elem(), to, seen)
	case reflect.Slice:
		for i := range v.Len() {
			redirectSchemaRefs(v.Index(i), to, seen)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			redirectSchemaRefs(v.MapIndex(key), to, seen)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				redirectSchemaRefs(v.Field(i), to, seen)
			}
		}
	}
}
