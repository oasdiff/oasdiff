package checker

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// requestTypeFormatBreaking reports whether a request type/format change is
// breaking. Removing the type constraint entirely is a non-breaking
// generalization (the server then accepts any value); otherwise defer to the
// direction-agnostic core.
func requestTypeFormatBreaking(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff) bool {
	if isTypeConstraintRemoved(typeDiff, schemaDiff) {
		return false
	}
	return typeOrFormatBreaking(typeDiff, formatDiff, isStronglyTyped(mediaType), schemaDiff)
}

// responseTypeFormatBreaking reports whether a response type/format change is
// breaking. Responses are covariant, so it is the same core check with the
// base/revision direction reversed.
func responseTypeFormatBreaking(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff) bool {
	typeDiff, formatDiff = reversed(typeDiff, formatDiff)
	return typeOrFormatBreaking(typeDiff, formatDiff, isStronglyTyped(mediaType), schemaDiff)
}

// reversed swaps the base/revision direction of the type and format diffs
// (Added<->Deleted, From<->To). A nil diff has no direction to reverse and is
// returned unchanged.
func reversed(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff) (*diff.StringsDiff, *diff.ValueDiff) {
	if typeDiff != nil {
		typeDiff = &diff.StringsDiff{Added: typeDiff.Deleted, Deleted: typeDiff.Added}
	}
	if formatDiff != nil {
		formatDiff = &diff.ValueDiff{From: formatDiff.To, To: formatDiff.From}
	}
	return typeDiff, formatDiff
}

// isTypeConstraintRemoved reports whether the type changed and the revision has
// no type, i.e. the type constraint was removed entirely.
func isTypeConstraintRemoved(typeDiff *diff.StringsDiff, schemaDiff *diff.SchemaDiff) bool {
	if typeDiff == nil {
		return false
	}
	rev := schemaDiff.Revision.Type
	return rev == nil || len(*rev) == 0
}

// typeOrFormatBreaking reports whether a type change or a format change is
// breaking. The two axes are evaluated independently: the change is breaking if
// either the type or the format is breaking on its own.
// stronglyTyped reflects the media type (see isStronglyTyped); callers that
// can't resolve it (request parameters) pass it explicitly.
func typeOrFormatBreaking(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, stronglyTyped bool, schemaDiff *diff.SchemaDiff) bool {
	typeBreaking := typeDiff != nil && !isTypeContained(typeDiff.Added, typeDiff.Deleted, stronglyTyped)
	formatBreaking := formatDiff != nil && !isFormatContained(schemaDiff.Revision.Type, formatDiff.To, formatDiff.From)
	return typeBreaking || formatBreaking
}

/*
isTypeContained checks if type2 is contained in type1
note that we don't support multiple types currenty
*/
func isTypeContained(to, from []string, stronglyTyped bool) bool {

	if len(to) == 1 && to[0] == "number" && len(from) == 1 && from[0] == "integer" {
		return true
	}

	// anything can be changed to string, unless it's "strongly typed"
	if !stronglyTyped {
		return len(to) == 0 || (len(to) == 1 && to[0] == "string")
	}

	return false
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
	stronglyTyped := typeOrFormatBreaking(typeDiff, formatDiff, true, schemaDiff)
	nonStronglyTyped := typeOrFormatBreaking(typeDiff, formatDiff, false, schemaDiff)

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

/*
isStronglyTyped checks if the media type is strongly typed, for example:
in text format, all numbers can also be interpreted as strings (1 can be a number or a string)
but in json, a number (1) is not the same as a string ("1")
*/
func isStronglyTyped(mediaType string) bool {
	return isJsonMediaType(mediaType)
}

func isJsonMediaType(mediaType string) bool {
	// Structured Syntax Suffixes: https://www.rfc-editor.org/rfc/rfc6838#section-4.2.8
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

// isFormatContained checks if from is contained in to
func isFormatContained(revisionType *openapi3.Types, to, from any) bool {

	// Removing a format constraint is a generalization (non-breaking), whatever
	// the type, including when the type was removed too (revisionType is nil).
	// Checked before the type switch so it applies for any revision type.
	if to == "" {
		return true
	}

	if revisionType == nil || len(*revisionType) > 1 {
		return false
	}

	// we don't support multiple types currenty, so just take the first one
	switch getSingleType(revisionType) {
	case "number":
		return to == "double" && from == "float"
	case "integer":
		return (to == "int64" && from == "int32") ||
			(to == "bigint" && from == "int32") ||
			(to == "bigint" && from == "int64")
	case "string":
		return (to == "date-time" && from == "date") ||
			(to == "date-time" && from == "time")
	}

	return false
}

func getSingleType(types *openapi3.Types) string {
	if types == nil || len(*types) == 0 {
		return ""
	}

	return (*types)[0]
}

func getBaseType(schemaDiff *diff.SchemaDiff) []string {
	return schemaDiff.Base.Type.Slice()
}

func getRevisionType(schemaDiff *diff.SchemaDiff) []string {
	return schemaDiff.Revision.Type.Slice()
}

func getBaseFormat(schemaDiff *diff.SchemaDiff) string {
	return schemaDiff.Base.Format
}

func getRevisionFormat(schemaDiff *diff.SchemaDiff) string {
	return schemaDiff.Revision.Format
}
