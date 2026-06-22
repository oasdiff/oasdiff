package diff

import (
	"slices"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// OneOfWrappingDiff describes a schema being wrapped into a oneOf of
// alternatives: a concrete object schema (properties, no oneOf) on the base side
// becomes a oneOf of object subschemas on the revision side, with no top-level
// type or properties. The base properties move into the alternatives rather than
// being removed, and a base-required property that is not required in every
// alternative becomes optional (a client may omit it in some valid branch).
//
// It follows the precedent of ListOfTypesDiff, which promotes a single<->oneOf
// transition to a first-class diff so the checker can avoid naive field-level
// false positives. The two are complementary, not reflections: ListOfTypesDiff
// handles the scalar type-set case (oneOf alternatives that differ by type, and
// its detector skips schemas with properties), while OneOfWrappingDiff handles
// the object case it skips (alternatives that are objects differing by
// properties/required). The checker reads it to suppress the spurious "property
// removed" findings the raw property diff would otherwise produce, and to report
// "became optional" instead. See oasdiff/oasdiff#702.
type OneOfWrappingDiff struct {
	// NumAlternatives is the number of oneOf alternatives on the revision side.
	NumAlternatives int `json:"numAlternatives,omitempty" yaml:"numAlternatives,omitempty"`
	// MovedProperties are base property names that appear in at least one
	// alternative, i.e. they moved into the wrapping rather than being removed.
	MovedProperties []string `json:"movedProperties,omitempty" yaml:"movedProperties,omitempty"`
	// RequiredBecameOptional are base-required properties that are not required
	// in every alternative.
	RequiredBecameOptional []string `json:"requiredBecameOptional,omitempty" yaml:"requiredBecameOptional,omitempty"`
}

// Empty indicates whether a change was found in this element.
func (diff *OneOfWrappingDiff) Empty() bool {
	return diff == nil ||
		(diff.NumAlternatives == 0 && len(diff.MovedProperties) == 0 && len(diff.RequiredBecameOptional) == 0)
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
		if ref == nil || ref.Value == nil {
			return nil // an unresolved alternative; don't guess
		}
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

	effectiveRequired := requiredIntersection(alts)
	becameOptional := make([]string, 0)
	for _, name := range base.Required {
		if slices.Contains(moved, name) && !slices.Contains(effectiveRequired, name) {
			becameOptional = append(becameOptional, name)
		}
	}

	sort.Strings(moved)
	sort.Strings(becameOptional)
	return &OneOfWrappingDiff{
		NumAlternatives:        len(alts),
		MovedProperties:        moved,
		RequiredBecameOptional: becameOptional,
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

// requiredIntersection returns the property names that are required in every
// alternative.
func requiredIntersection(alts []*openapi3.Schema) []string {
	if len(alts) == 0 {
		return nil
	}
	counts := map[string]int{}
	for _, alt := range alts {
		for _, r := range slices.Compact(slices.Clone(sortedStrings(alt.Required))) {
			counts[r]++
		}
	}
	out := make([]string, 0)
	for name, c := range counts {
		if c == len(alts) {
			out = append(out, name)
		}
	}
	return out
}

// sortedStrings returns a sorted copy, so slices.Compact can dedupe.
func sortedStrings(in []string) []string {
	out := slices.Clone(in)
	sort.Strings(out)
	return out
}
