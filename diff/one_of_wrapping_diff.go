package diff

import (
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// OneOfWrappingDiff marks that a concrete object schema was wrapped into a
// oneOf of alternatives: a concrete object schema (properties, no oneOf) on the
// base side becomes a oneOf of object subschemas on the revision side, with no
// top-level type or properties. The base properties move into the alternatives
// rather than being removed.
//
// Wrapping a concrete object request body into a oneOf is a breaking
// restructuring: under oneOf (validate against exactly one), a previously valid
// payload can match multiple overlapping alternatives and be rejected.
//
// It follows the precedent of ListOfTypesDiff, which promotes a single<->oneOf
// transition to a first-class diff so the checker can avoid naive field-level
// false positives. The two are complementary, not reflections: ListOfTypesDiff
// handles the scalar type-set case (oneOf alternatives that differ by type, and
// its detector skips schemas with properties), while OneOfWrappingDiff handles
// the object case it skips (alternatives that are objects differing by
// properties/required). The checker reads it to (a) suppress the spurious
// "property removed" findings the raw property diff would otherwise produce for
// the moved properties, and (b) emit an accurate breaking finding. See
// oasdiff/oasdiff#702.
type OneOfWrappingDiff struct {
	// NumAlternatives is the number of oneOf alternatives on the revision side.
	NumAlternatives int `json:"numAlternatives,omitempty" yaml:"numAlternatives,omitempty"`
	// MovedProperties are base property names that appear in at least one
	// alternative, i.e. they moved into the wrapping rather than being removed.
	MovedProperties []string `json:"movedProperties,omitempty" yaml:"movedProperties,omitempty"`
}

// Empty indicates whether a change was found in this element.
func (diff *OneOfWrappingDiff) Empty() bool {
	return diff == nil ||
		(diff.NumAlternatives == 0 && len(diff.MovedProperties) == 0)
}

// getOneOfWrappingDiff detects the "oneOf wrapping" pattern (base is a concrete
// object schema, revision wraps equivalent alternatives in a oneOf) and returns
// nil when it doesn't apply. The reverse (unwrapping a oneOf into a concrete
// schema) is not detected here.
func getOneOfWrappingDiff(base, revision *openapi3.Schema) *OneOfWrappingDiff {
	if base == nil || revision == nil {
		return nil
	}
	// base: a concrete object (has properties, no oneOf of its own).
	if len(base.OneOf) > 0 || len(base.Properties) == 0 {
		return nil
	}
	// revision: a oneOf wrapper with no top-level type or properties.
	if len(revision.OneOf) == 0 || len(revision.Properties) > 0 || len(revision.Type.Slice()) > 0 {
		return nil
	}

	alts := make([]*openapi3.Schema, 0, len(revision.OneOf))
	for _, ref := range revision.OneOf {
		alts = append(alts, ref.Value)
	}

	moved := make([]string, 0, len(base.Properties))
	for name := range base.Properties {
		if propertyInAnyAlternative(name, alts) {
			moved = append(moved, name)
		}
	}
	if len(moved) == 0 {
		return nil
	}

	sort.Strings(moved)
	return &OneOfWrappingDiff{
		NumAlternatives: len(alts),
		MovedProperties: moved,
	}
}

// propertyInAnyAlternative reports whether the named property exists in at least
// one of the oneOf alternatives.
func propertyInAnyAlternative(name string, alts []*openapi3.Schema) bool {
	for _, alt := range alts {
		if _, ok := alt.Properties[name]; ok {
			return true
		}
	}
	return false
}
