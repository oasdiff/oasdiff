package allof

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// MergeSpec merges all instances of allOf in place, across every schema in
// the document, and gives every cycle the merge anchors a $ref
// (nameAnchoredCycles), so the merged document always serializes. Merge
// handles each schema's whole subtree (including its own cycle tracking), so
// the walk hands it each attachment point once and skips descent.
func MergeSpec(spec *openapi3.T) (*openapi3.T, error) {
	var anchored []*openapi3.SchemaRef
	hints := map[*openapi3.Schema]string{}
	err := spec.WalkSchemas(func(_ string, s *openapi3.SchemaRef) error {
		m, edges, mergeHints, err := mergeWithAnchors(*s)
		if err != nil {
			return err
		}
		anchored = append(anchored, edges...)
		maps.Copy(hints, mergeHints)
		// Every $ref to this schema shares one Value, so writing the merge
		// into it updates every use. Assigning s.Value would update only
		// this reference and leave the rest unmerged.
		*s.Value = *m
		// The copy leaves the subtree's self-references aimed at the object
		// Merge returned rather than the one just written into: a recursive
		// schema's one-object cycle becomes a two-object cycle, and later
		// attachment points then see two distinct originals for one schema,
		// which the merge cache cannot unify.
		redirectSchemaRefs(reflect.ValueOf(s.Value), m, s.Value, map[*openapi3.Schema]bool{})
		// the copy also detaches m from its hint; the written-back object is
		// the one the anchored edges now point at
		if hint, ok := hints[m]; ok {
			hints[s.Value] = hint
		}
		return openapi3.SkipSubtree
	})
	if err != nil {
		return spec, err
	}
	nameAnchoredCycles(spec, anchored, hints)
	return spec, nil
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
// whose Value is from at to instead. The traversal is type-driven so a new
// schema field carrying subschemas is covered without being listed here.
func redirectSchemaRefs(v reflect.Value, from, to *openapi3.Schema, seen map[*openapi3.Schema]bool) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		if ref, ok := v.Interface().(*openapi3.SchemaRef); ok {
			if ref.Value == from {
				ref.Value = to
			}
			redirectSchemaRefs(reflect.ValueOf(ref.Value), from, to, seen)
			return
		}
		if schema, ok := v.Interface().(*openapi3.Schema); ok {
			if schema == nil || seen[schema] {
				return
			}
			seen[schema] = true
		}
		redirectSchemaRefs(v.Elem(), from, to, seen)
	case reflect.Slice:
		for i := range v.Len() {
			redirectSchemaRefs(v.Index(i), from, to, seen)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			redirectSchemaRefs(v.MapIndex(key), from, to, seen)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				redirectSchemaRefs(v.Field(i), from, to, seen)
			}
		}
	}
}
