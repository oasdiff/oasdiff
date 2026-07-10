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

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
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

						result = append(result, opInfo.NewApiChange(
							id,
							[]any{paramLocation, paramName, getTypeFormatDimension(schemaDiff), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff)},
							comment,
						).WithSchema(schemaDiff).WithSources(baseSource, revisionSource))
					}

					checkModifiedPropertiesDiff(
						schemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {

							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "type")
							schemaDiff := propertyDiff
							typeDiff := schemaDiff.TypeDiff
							formatDiff := schemaDiff.FormatDiff

							if !typeDiff.Empty() || !formatDiff.Empty() {

								id, comment := checkRequestParameterPropertyTypeChanged(typeDiff, formatDiff, schemaDiff)

								result = append(result, opInfo.NewApiChange(
									id,
									[]any{paramLocation, paramName, getTypeFormatDimension(schemaDiff), propertyFullName(propertyPath, propertyName), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff)},
									comment,
								).WithSchema(schemaDiff).WithSources(propBaseSource, propRevisionSource))
							}
						})
				}
			}
		}
	}
	return result
}

/*
checkRequestParameterPropertyTypeChanged checks the level of the change in the request parameter property type
Explanation:
Objects can be passed in the request parameters, for example, the following calls are equivalent:
PHP style: GET http://localhost:8080/api/tickets?params[id]=123&params[color]=green
JSON: GET http://localhost:8080/api/tickets?params={"id":"123","color":"green"}

The "params" object has two properties: "id" and "color", both with type "string", but note that the "id" values are actually numbers.
Imagine that the OpenAPI type of property "id" was changed from "number" to "string".
In the first example, the change is non-breaking, because the PHP format for numbers and strings is the same: we refer to this as non-strongly-typed.
But in the second example, the change is breaking, because the JSON format requires quotes for strings: we refer to this as strongly-typed.

This is the only request type location that forks three ways
(generalized / specialized / changed-as-a-warning). The other request type
locations resolve strong-vs-non-strong definitively (the body and body
properties from a known media type; a scalar parameter is always a string on
the wire), so a binary generalized/changed verdict is correct there. Only an
object parameter's serialization is unknown here, so when the two verdicts
disagree we can't be sure it's breaking and report a warning.
*/
func checkRequestParameterPropertyTypeChanged(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, schemaDiff *diff.SchemaDiff) (string, string) {

	// since we don't know if the object is strogly-typed or not, we check both
	stronglyTyped := typeOrFormatBreaking(typeDiff, formatDiff, true, schemaDiff.Revision.Type)
	nonStronglyTyped := typeOrFormatBreaking(typeDiff, formatDiff, false, schemaDiff.Revision.Type)

	// if strongly-typed and non-strongly-typed don't agree, it's a warning since we can't be sure that it's breaking
	if stronglyTyped != nonStronglyTyped {
		return RequestParameterPropertyTypeChangedId, RequestParameterPropertyTypeChangedCommentId
	}

	// if both are breaking it's an error
	if stronglyTyped {
		return RequestParameterPropertyTypeSpecializedId, ""
	}

	// if neither are breaking it's an informational change
	return RequestParameterPropertyTypeGeneralizedId, ""
}
