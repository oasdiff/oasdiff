package checker

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterTypeChangedId                = "request-parameter-type-changed"
	RequestParameterTypeGeneralizedId            = "request-parameter-type-generalized"
	RequestParameterPropertyTypeChangedId        = "request-parameter-property-type-changed"
	RequestParameterPropertyTypeGeneralizedId    = "request-parameter-property-type-generalized"
	RequestParameterPropertyTypeSpecializedId    = "request-parameter-property-type-specialized"
	RequestParameterPropertyTypeChangedCommentId = "request-parameter-property-type-changed-warn-comment"
)

// isParameterScalarToFormExplodeArray reports whether typeDiff describes a
// scalar-to-array change on a parameter using form/explode serialization
// (the default for query and cookie parameters), where the new array's
// element type matches the old scalar type. Per the OpenAPI spec, a client
// sending a single value (?color=red) is a valid one-element array under
// these serialization rules, so the schema change is backwards-compatible.
func isParameterScalarToFormExplodeArray(paramDiff *diff.ParameterDiff, typeDiff *diff.StringsDiff) bool {
	if paramDiff == nil || paramDiff.Revision == nil || paramDiff.SchemaDiff == nil {
		return false
	}
	if typeDiff == nil {
		return false
	}

	// Compare modulo "null". Under OpenAPI 3.1 the type can be a JSON-Schema
	// array, so a nullable scalar is ["string","null"] and a nullable array is
	// ["array","null"]. Preserving or adding nullability around a
	// scalar-to-form/explode-array widening is still backwards-compatible, so
	// strip "null" from both the type diff and the item types before applying
	// the 3.0 single-type equality. Multi-non-null types and item-type
	// mismatches still fail the len==1 / equality checks, so the relaxation
	// never declares an unsafe change safe.
	deletedSansNull := withoutNull(typeDiff.Deleted)
	addedSansNull := withoutNull(typeDiff.Added)
	if len(addedSansNull) != 1 || addedSansNull[0] != "array" ||
		len(deletedSansNull) != 1 || deletedSansNull[0] == "array" {
		return false
	}

	method, err := paramDiff.Revision.SerializationMethod()
	if err != nil || method == nil ||
		method.Style != openapi3.SerializationForm || !method.Explode {
		return false
	}

	revSchema := paramDiff.SchemaDiff.Revision
	if revSchema == nil || revSchema.Items == nil || revSchema.Items.Value == nil {
		return false
	}
	itemTypes := withoutNull(revSchema.Items.Value.Type.Slice())
	if len(itemTypes) != 1 || itemTypes[0] != deletedSansNull[0] {
		return false
	}

	// A matching item type is not enough: the item must not constrain values more
	// than the base scalar did, or a value valid under the base could be rejected
	// by the item (e.g. base `string` -> item `string` with a `pattern` rejects
	// "5"). Otherwise the relaxation would declare a breaking change safe (#1024).
	return itemNoStricterThanBaseScalar(paramDiff.SchemaDiff.Base, revSchema.Items.Value)
}

// itemNoStricterThanBaseScalar reports whether the array's item schema accepts
// every value the base scalar accepted. The item may DROP the base scalar's
// value constraints (that only widens the accepted set) but must not ADD or
// CHANGE one, which could reject a value valid under the base (e.g. base
// `string` -> item `string` with a `pattern` rejects "5"). So for each value
// constraint the item is acceptable only when it is unset, or set identically to
// the base; anything else is conservatively treated as not-provably-safe, so the
// scalar-to-array widening is never declared safe when the items could narrow.
func itemNoStricterThanBaseScalar(base, item *openapi3.Schema) bool {
	if base == nil || item == nil {
		return false
	}
	if item.Pattern != "" && item.Pattern != base.Pattern {
		return false
	}
	if item.Format != "" && item.Format != base.Format {
		return false
	}
	if item.MinLength != 0 && item.MinLength != base.MinLength {
		return false
	}
	if len(item.Enum) != 0 && !reflect.DeepEqual(item.Enum, base.Enum) {
		return false
	}
	if item.MaxLength != nil && !reflect.DeepEqual(item.MaxLength, base.MaxLength) {
		return false
	}
	if item.Min != nil && !(reflect.DeepEqual(item.Min, base.Min) && item.ExclusiveMin == base.ExclusiveMin) {
		return false
	}
	if item.Max != nil && !(reflect.DeepEqual(item.Max, base.Max) && item.ExclusiveMax == base.ExclusiveMax) {
		return false
	}
	if item.MultipleOf != nil && !reflect.DeepEqual(item.MultipleOf, base.MultipleOf) {
		return false
	}
	// A composed item (allOf/anyOf/oneOf/not) is not a plain scalar; cannot prove
	// it accepts all base values, so treat it as not-provably-safe.
	return len(item.AllOf) == 0 && len(item.AnyOf) == 0 && len(item.OneOf) == 0 && item.Not == nil
}

// withoutNull returns the type list with the JSON-Schema "null" type removed,
// so OpenAPI 3.1 nullable types (e.g. ["string","null"]) compare by their
// non-null type.
func withoutNull(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		if t != "null" {
			out = append(out, t)
		}
	}
	return out
}

func RequestParameterTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ParametersDiff == nil {
				continue
			}

			for paramLocation, paramDiffs := range operationItem.ParametersDiff.Modified {
				for paramName, paramDiff := range paramDiffs {
					if paramDiff.SchemaDiff == nil {
						continue
					}

					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, "type")
					schemaDiff := paramDiff.SchemaDiff
					typeDiff := schemaDiff.TypeDiff
					formatDiff := schemaDiff.FormatDiff

					if !typeDiff.Empty() || !formatDiff.Empty() {

						// Suppress single<->oneOf/anyOf transitions (handled by the list-of-types checker)
						if shouldSuppressTypeChangedForListOfTypes(schemaDiff) {
							continue
						}

						// Suppress null-only type changes (handled by nullable checkers)
						if isNullTypeChange(typeDiff) && formatDiff.Empty() {
							continue
						}

						id := RequestParameterTypeGeneralizedId

						// The parameter's own value is serialized as a string on the wire
						// (query/path/header/cookie), so it is non-strongly-typed: a binary
						// generalized/changed verdict with stronglyTyped=false. This differs
						// on purpose from the property-level check below, which cannot tell
						// how an object parameter is serialized and therefore forks three ways.
						if typeOrFormatBreaking(typeDiff, formatDiff, false, schemaDiff.Revision.Type) &&
							!isParameterScalarToFormExplodeArray(paramDiff, typeDiff) {
							id = RequestParameterTypeChangedId
						}

						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff)},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}

					CheckModifiedPropertiesDiff(
						schemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {

							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "type")
							schemaDiff := propertyDiff
							typeDiff := schemaDiff.TypeDiff
							formatDiff := schemaDiff.FormatDiff

							if !typeDiff.Empty() || !formatDiff.Empty() {

								// Suppress single<->oneOf/anyOf transitions (handled by the list-of-types checker)
								if shouldSuppressPropertyTypeChangedForListOfTypes(schemaDiff) {
									return
								}

								// Suppress null-only type changes (handled by nullable checkers)
								if isNullTypeChange(typeDiff) && formatDiff.Empty() {
									return
								}

								id, comment := checkRequestParameterPropertyTypeChanged(typeDiff, formatDiff, schemaDiff)

								result = append(result, NewApiChange(
									id,
									config,
									[]any{paramLocation, paramName, propertyFullName(propertyPath, propertyName), getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff)},
									comment,
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource))
							}
						})
				}
			}
		}
	}
	return result
}
