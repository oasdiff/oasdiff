package checker

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/utils"
)

// breakingTypeFormatChangedInResponseProperty checks if the type or format of a response property was changed in a breaking way
func breakingTypeFormatChangedInResponseProperty(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff) bool {

	if typeDiff != nil {
		typeDiff = &diff.StringsDiff{
			Added:   typeDiff.Deleted,
			Deleted: typeDiff.Added,
		}
	}

	if formatDiff != nil {
		formatDiff = &diff.ValueDiff{
			From: formatDiff.To,
			To:   formatDiff.From,
		}
	}

	return breakingTypeFormatChangedInRequestProperty(typeDiff, formatDiff, mediaType, schemaDiff)
}

// breakingTypeFormatChangedInRequestProperty checks if the type or format of a request property was changed in a breaking way
func breakingTypeFormatChangedInRequestProperty(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff) bool {
	return breakingTypeFormatChangedInRequest(typeDiff, formatDiff, isStronglyTyped(mediaType), schemaDiff)
}

// breakingTypeFormatChangedInRequest checks if the type or format of a request was changed in a breaking way
func breakingTypeFormatChangedInRequest(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, stronglyTyped bool, schemaDiff *diff.SchemaDiff) bool {

	if typeDiff != nil {
		return !isTypeContained(typeDiff.Added, typeDiff.Deleted, stronglyTyped)
	}

	if formatDiff != nil {
		return !isFormatContained(schemaDiff.Revision.Type, formatDiff.To, formatDiff.From)
	}

	return false
}

/*
isTypeContained checks if type2 is contained in type1
note that we don't support multiple types currenty
*/
func isTypeContained(to, from utils.StringList, stronglyTyped bool) bool {

	if to.Is("number") && from.Is("integer") {
		return true
	}

	// anything can be changed to string, unless it's "strongly typed"
	if !stronglyTyped {
		return to.Empty() || to.Is("string")
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
*/
func checkRequestParameterPropertyTypeChanged(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, schemaDiff *diff.SchemaDiff) (string, string) {

	// since we don't know if the object is strogly-typed or not, we check both
	stronglyTyped := breakingTypeFormatChangedInRequest(typeDiff, formatDiff, true, schemaDiff)
	nonStronglyTyped := breakingTypeFormatChangedInRequest(typeDiff, formatDiff, false, schemaDiff)

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
func isFormatContained(revisionType *openapi3.Types, to, from interface{}) bool {

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
		return (to == "date-time" && from == "date" ||
			to == "date-time" && from == "time")
	}

	return false
}

func getSingleType(types *openapi3.Types) string {
	if types == nil || len(*types) == 0 {
		return ""
	}

	return (*types)[0]
}

func getBaseType(schemaDiff *diff.SchemaDiff) utils.StringList {
	return schemaDiff.Base.Type.Slice()
}

func getRevisionType(schemaDiff *diff.SchemaDiff) utils.StringList {
	return schemaDiff.Revision.Type.Slice()
}

func getBaseFormat(schemaDiff *diff.SchemaDiff) string {
	return schemaDiff.Base.Format
}

func getRevisionFormat(schemaDiff *diff.SchemaDiff) string {
	return schemaDiff.Revision.Format
}
