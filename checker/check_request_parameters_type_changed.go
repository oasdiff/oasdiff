package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterTypeChangedId                 = "request-parameter-type-changed"
	RequestParameterTypeGeneralizedId             = "request-parameter-type-generalized"
	RequestParameterPropertyTypeChangedId         = "request-parameter-property-type-changed"
	RequestParameterPropertyTypeGeneralizedId     = "request-parameter-property-type-generalized"
	RequestParameterPropertyTypeSpecializedId     = "request-parameter-property-type-specialized"
	RequestParameterPropertyTypeChangedCommentId  = "request-parameter-property-type-changed-warn-comment"
	RequestParameterTypeFormExplodeArrayCommentId = "request-parameter-type-form-explode-array-comment"
)

// isParameterScalarToFormExplodeArray reports whether typeDiff describes a
// non-array-to-array change on a parameter using form/explode serialization
// (the default for query and cookie parameters), where the new array's item
// accepts every value the old scalar did. Per the OpenAPI spec, a client
// sending a single value (?color=red) is a valid one-element array under these
// serialization rules, so such a widening is backwards-compatible.
func isParameterScalarToFormExplodeArray(paramDiff *diff.ParameterDiff, typeDiff *diff.StringsDiff) bool {
	if paramDiff == nil || paramDiff.Revision == nil || paramDiff.SchemaDiff == nil ||
		paramDiff.SchemaDiff.Base == nil || paramDiff.SchemaDiff.Revision == nil {
		return false
	}
	if typeDiff == nil {
		return false
	}

	// The revision must be a (possibly nullable) array and the base a non-array
	// scalar, or union of scalars. "null" is stripped so that preserving or adding
	// nullability around the widening stays in scope (a nullable scalar is
	// ["string","null"], a nullable array ["array","null"]). A base that is
	// untyped or already an array is out of scope and stays on the breaking path.
	addedSansNull := withoutNull(typeDiff.Added)
	baseTypes := withoutNull(paramDiff.SchemaDiff.Base.Type.Slice())
	if len(addedSansNull) != 1 || addedSansNull[0] != "array" || len(baseTypes) == 0 {
		return false
	}

	method, err := paramDiff.Revision.SerializationMethod()
	if err != nil || method == nil ||
		method.Style != openapi3.SerializationForm || !method.Explode {
		return false
	}

	revSchema := paramDiff.SchemaDiff.Revision
	if revSchema.Items == nil || revSchema.Items.Value == nil {
		return false
	}
	return itemAcceptsBaseScalarValues(paramDiff.SchemaDiff.Base, revSchema.Items.Value)
}

// itemAcceptsBaseScalarValues reports whether the array's item schema accepts
// every value the base scalar accepted, so wrapping the scalar in the array is
// safe. It splits the question along the two axes the rest of check_types.go
// uses:
//
//   - Type: the item's type must contain the base scalar's type(s). Query and
//     cookie parameters are weakly typed (every value is a string on the wire),
//     so isTypeContained applies the "anything is accepted as a string" rule and
//     the integer-within-number lattice. This is what makes, for example,
//     ["string","integer"] -> array<string> safe, consistent with how the
//     scalar-to-scalar ["string","integer"] -> string change is already treated.
//   - Value constraints: the item must not add or tighten a non-type constraint,
//     which could reject a value valid under the base (e.g. a `pattern` that
//     rejects "5", #1024). With the type removed (compared above), the diff
//     engine's validation-equivalence check covers every remaining keyword
//     (including `const` and the 3.1 conditionals) centrally, ignoring
//     annotation-only differences such as description.
func itemAcceptsBaseScalarValues(base, item *openapi3.Schema) bool {
	if base == nil || item == nil {
		return false
	}

	if !isTypeContained(withoutNull(item.Type.Slice()), withoutNull(base.Type.Slice()), false) {
		return false
	}

	// Compare the non-type contract. Zero the type (handled above) and the
	// nullability (handled via "null" stripping / the nullable checkers) on copies
	// so the equivalence check sees only the value constraints.
	baseValue, itemValue := *base, *item
	baseValue.Type, itemValue.Type = nil, nil
	baseValue.Nullable, itemValue.Nullable = false, false
	return diff.SchemaRefsValidationEquivalent(diff.NewConfig(),
		&openapi3.SchemaRef{Value: &baseValue}, &openapi3.SchemaRef{Value: &itemValue})
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
						comment := ""

						// The parameter's own value is serialized as a string on the wire
						// (query/path/header/cookie), so it is non-strongly-typed: a binary
						// generalized/changed verdict with stronglyTyped=false. This differs
						// on purpose from the property-level check below, which cannot tell
						// how an object parameter is serialized and therefore forks three ways.
						if typeOrFormatBreaking(typeDiff, formatDiff, false, schemaDiff.Revision.Type) {
							if isParameterScalarToFormExplodeArray(paramDiff, typeDiff) {
								// The type change would otherwise be breaking; explain why
								// widening to a form/explode array is safe so the verdict is
								// not surprising.
								comment = RequestParameterTypeFormExplodeArrayCommentId
							} else {
								id = RequestParameterTypeChangedId
							}
						}

						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, getTypeFormatDimension(schemaDiff), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff)},
							comment,
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
									[]any{paramLocation, paramName, getTypeFormatDimension(schemaDiff), propertyFullName(propertyPath, propertyName), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff)},
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
