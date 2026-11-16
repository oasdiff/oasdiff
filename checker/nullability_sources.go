package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

// nullabilitySources locates a nullability transition in the base and the
// revision spec. Each form nullabilityChange recognizes declares nullability
// under a different keyword, so the keyword looked up follows the form:
//
//   - the nullable keyword: nullable
//   - a "null" entry in the type array: that entry, or type on the side
//     without one
//   - the nullable oneOf wrapping: oneOf on the wrapped side, type on the other
//   - a oneOf or anyOf whose branches decide: the branch whose own nullability
//     changed, since that is the line that was edited, or the composition
//     keyword when no single branch changed
//
// A side that does not have the keyword is located at the schema itself.
func nullabilitySources(operationsSources diff.OperationsSourcesMap, operationItem *diff.MethodDiff, d *diff.SchemaDiff) (*Source, *Source) {
	if d == nil || !hasOrigin(d.Base) && !hasOrigin(d.Revision) {
		return location.SchemaFieldSources(operationsSources, operationItem, d, "nullable")
	}

	var keywords []string
	switch {
	case d.NullableDiff != nil:
		keywords = []string{"nullable"}
	case d.Base == nil || d.Revision == nil:
	case nullRemovedFromTypeArray(d.TypeDiff, d.Revision.Type),
		nullAddedToTypeArray(d.TypeDiff, d.Base.Type),
		isNullableWrapping(d),
		isNullableUnwrapping(d):
		keywords = []string{"type", "oneOf"}
	default:
		if branch := branchThatChangedNullability(d); branch != nil {
			return nullabilitySources(operationsSources, operationItem, branch)
		}
		keywords = []string{"oneOf", "anyOf"}
	}

	return nullabilitySource(operationsSources, operationItem.Base, d.Base, keywords),
		nullabilitySource(operationsSources, operationItem.Revision, d.Revision, keywords)
}

func hasOrigin(schema *openapi3.Schema) bool {
	return schema != nil && schema.Origin != nil
}

// nullabilitySource locates the first of the keywords the schema has, and for
// type, the "null" entry of the type array when there is one.
func nullabilitySource(operationsSources diff.OperationsSourcesMap, operation *openapi3.Operation, schema *openapi3.Schema, keywords []string) *Source {
	if operation == nil || !hasOrigin(schema) {
		return nil
	}
	if slices.Contains(keywords, "type") {
		if source := location.NewSourceFromSequenceItem(operationsSources, operation, schema.Origin, "type", "null"); source != nil {
			return source
		}
	}
	return location.SchemaKeywordSource(operationsSources, operation, schema, keywords...)
}

// branchThatChangedNullability returns the oneOf or anyOf branch whose own
// nullability changed, or nil when there is not exactly one: the composition's
// nullability then moved because branches were added or removed, or because
// several changed at once.
func branchThatChangedNullability(d *diff.SchemaDiff) *diff.SchemaDiff {
	var result *diff.SchemaDiff
	for _, subschemas := range []*diff.SubschemasDiff{d.OneOfDiff, d.AnyOfDiff} {
		if subschemas == nil {
			continue
		}
		for _, modified := range subschemas.Modified {
			if modified == nil {
				continue
			}
			if transition, _ := nullabilityChange(modified.Diff); transition == nullabilityUnchanged {
				continue
			}
			if result != nil {
				return nil
			}
			result = modified.Diff
		}
	}
	return result
}
