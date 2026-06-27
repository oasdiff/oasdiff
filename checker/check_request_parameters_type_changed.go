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

	// A matching item type is not enough: the widening is safe only when the item
	// schema accepts every value the base scalar accepted. We prove that the
	// narrow way - by requiring the item to be the base scalar with nothing changed
	// but the array wrapping. Any other difference (a tighter or different value
	// constraint, e.g. base `string` -> item `string` with a `pattern` that rejects
	// "5") fails the comparison, so the relaxation is never declared safe when the
	// items could narrow (#1024).
	return itemMatchesBaseScalar(paramDiff.SchemaDiff.Base, revSchema.Items.Value)
}

// itemMatchesBaseScalar reports whether the array's item schema is the base
// scalar with nothing changed but the array wrapping, so it provably accepts
// exactly the values the base scalar accepted. It compares the two schemas'
// value-validation surface for equality: every validation keyword (including
// `const` and the OpenAPI 3.1 conditional keywords) is compared by zeroing only
// the type (handled separately above), nullability, and the non-validating
// annotation/metadata fields, then requiring deep equality of the rest.
//
// The comparison is deliberately exhaustive rather than a hand-listed set of
// constraints: a forgotten annotation field merely stays in the comparison and
// makes the check stricter (a safe over-report), whereas a forgotten *constraint*
// in an allow-list would silently declare a narrowing safe. So the residual risk
// is pushed to the harmless direction.
func itemMatchesBaseScalar(base, item *openapi3.Schema) bool {
	if base == nil || item == nil {
		return false
	}
	return reflect.DeepEqual(validationSurface(base), validationSurface(item))
}

// validationSurface returns a copy of s with the type, nullability, and all
// non-validating annotation/metadata fields zeroed, leaving only the keywords
// that constrain which values validate. The type is compared separately (scalar
// vs array, modulo "null"); nullability is treated as not affecting safety, in
// keeping with the null-stripping the caller already applies.
func validationSurface(s *openapi3.Schema) *openapi3.Schema {
	c := *s
	c.Type = nil
	c.Nullable = false

	// Non-validating metadata (these never reject a value).
	c.Origin = nil
	c.Extensions = nil
	c.Title = ""
	c.Description = ""
	c.Default = nil
	c.Example = nil
	c.Examples = nil
	c.ExternalDocs = nil
	c.Deprecated = false
	c.ReadOnly = false
	c.WriteOnly = false
	c.XML = nil
	c.Discriminator = nil
	c.Comment = ""
	c.SchemaDialect = ""
	c.SchemaID = ""
	c.Anchor = ""
	c.DynamicRef = ""
	c.DynamicAnchor = ""
	c.Defs = nil
	return &c
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
