package checker

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// Core logic functions for ListOfTypes checking that can be reused by both full checkers and suppression functions

// checkPropertyListOfTypesChange checks if a property change involves list-of-types patterns and returns breaking changes
func checkPropertyListOfTypesChange(opInfo opInfo, propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff,
	mediaType string, responseStatus string, isRequest bool) Changes {

	result := make(Changes, 0)

	if propertyDiff.ListOfTypesDiff.Empty() {
		return result
	}

	listDiff := propertyDiff.ListOfTypesDiff

	// Determine breaking changes based on variance rules
	var messageId string
	var args []any

	if isRequest {
		// Request properties are contravariant: adding types = non-breaking, removing types = breaking
		if len(listDiff.Deleted) > 0 {
			messageId = RequestPropertyListOfTypesNarrowedId
			args = []any{propertyFullName(propertyPath, propertyName), joinTypes(listDiff.Deleted), mediaType}
		} else {
			messageId = RequestPropertyListOfTypesWidenedId
			args = []any{propertyFullName(propertyPath, propertyName), joinTypes(listDiff.Added), mediaType}
		}
	} else {
		// Response properties are covariant: adding types = breaking, removing types = non-breaking
		if len(listDiff.Added) > 0 {
			messageId = ResponsePropertyListOfTypesWidenedId
			args = []any{propertyFullName(propertyPath, propertyName), joinTypes(listDiff.Added), mediaType, responseStatus}
		} else {
			messageId = ResponsePropertyListOfTypesNarrowedId
			args = []any{propertyFullName(propertyPath, propertyName), joinTypes(listDiff.Deleted), mediaType, responseStatus}
		}
	}

	baseSource, revisionSource := SchemaFieldSources(opInfo.operationsSources, opInfo.methodDiff, propertyDiff, "type")
	result = append(result, opInfo.NewApiChange(
		messageId,
		args,
		"",
	).WithSchema(propertyDiff).WithSources(baseSource, revisionSource))

	return result
}

// checkBodyListOfTypesChange checks if a request/response body change involves list-of-types patterns and returns breaking changes
func checkBodyListOfTypesChange(opInfo opInfo, schemaDiff *diff.SchemaDiff, mediaType string, responseStatus string,
	isRequest bool) Changes {

	result := make(Changes, 0)

	if schemaDiff.ListOfTypesDiff.Empty() {
		return result
	}

	listDiff := schemaDiff.ListOfTypesDiff

	// Determine breaking changes based on variance rules
	var messageId string
	var args []any

	if isRequest {
		// Request bodies are contravariant: adding types = non-breaking, removing types = breaking
		if len(listDiff.Deleted) > 0 {
			messageId = RequestBodyListOfTypesNarrowedId
			args = []any{joinTypes(listDiff.Deleted), mediaType}
		} else {
			messageId = RequestBodyListOfTypesWidenedId
			args = []any{joinTypes(listDiff.Added), mediaType}
		}
	} else {
		// Response bodies are covariant: adding types = breaking, removing types = non-breaking
		if len(listDiff.Added) > 0 {
			messageId = ResponseBodyListOfTypesWidenedId
			args = []any{joinTypes(listDiff.Added), mediaType, responseStatus}
		} else {
			messageId = ResponseBodyListOfTypesNarrowedId
			args = []any{joinTypes(listDiff.Deleted), mediaType, responseStatus}
		}
	}

	baseSource, revisionSource := SchemaFieldSources(opInfo.operationsSources, opInfo.methodDiff, schemaDiff, "type")
	result = append(result, opInfo.NewApiChange(
		messageId,
		args,
		"",
	).WithSchema(schemaDiff).WithSources(baseSource, revisionSource))

	return result
}

// checkParameterListOfTypesChange checks if a parameter change involves list-of-types patterns and returns breaking changes
func checkParameterListOfTypesChange(opInfo opInfo, paramDiff *diff.ParameterDiff, param *openapi3.Parameter) Changes {

	result := make(Changes, 0)

	if paramDiff == nil || paramDiff.SchemaDiff == nil || paramDiff.SchemaDiff.ListOfTypesDiff.Empty() {
		return result
	}

	listDiff := paramDiff.SchemaDiff.ListOfTypesDiff

	// Parameters are contravariant: adding types = non-breaking, removing types = breaking
	var messageId string
	var args []any

	if len(listDiff.Deleted) > 0 {
		messageId = RequestParameterListOfTypesNarrowedId
		args = []any{param.In, param.Name, joinTypes(listDiff.Deleted)}
	} else {
		messageId = RequestParameterListOfTypesWidenedId
		args = []any{param.In, param.Name, joinTypes(listDiff.Added)}
	}

	baseSource, revisionSource := SchemaFieldSources(opInfo.operationsSources, opInfo.methodDiff, paramDiff.SchemaDiff, "type")
	result = append(result, opInfo.NewApiChange(
		messageId,
		args,
		"",
	).WithSchema(paramDiff.SchemaDiff).WithSources(baseSource, revisionSource))

	return result
}

// checkParameterPropertyListOfTypesChange checks if a parameter property change involves list-of-types patterns and returns breaking changes
func checkParameterPropertyListOfTypesChange(opInfo opInfo, propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff,
	param *openapi3.Parameter) Changes {

	result := make(Changes, 0)

	if propertyDiff.ListOfTypesDiff.Empty() {
		return result
	}

	listDiff := propertyDiff.ListOfTypesDiff

	// Parameter properties are contravariant: adding types = non-breaking, removing types = breaking
	var messageId string
	var args []any

	if len(listDiff.Deleted) > 0 {
		messageId = RequestParameterPropertyListOfTypesNarrowedId
		args = []any{propertyFullName(propertyPath, propertyName), param.In, param.Name, joinTypes(listDiff.Deleted)}
	} else {
		messageId = RequestParameterPropertyListOfTypesWidenedId
		args = []any{propertyFullName(propertyPath, propertyName), param.In, param.Name, joinTypes(listDiff.Added)}
	}

	baseSource, revisionSource := SchemaFieldSources(opInfo.operationsSources, opInfo.methodDiff, propertyDiff, "type")
	result = append(result, opInfo.NewApiChange(
		messageId,
		args,
		"",
	).WithSchema(propertyDiff).WithSources(baseSource, revisionSource))

	return result
}

// Helper function to join types for display in messages
func joinTypes(types []string) string {
	if len(types) == 0 {
		return ""
	}
	if len(types) == 1 {
		return types[0]
	}

	var result strings.Builder
	for i, t := range types {
		if i > 0 {
			if i == len(types)-1 {
				result.WriteString(" and ")
			} else {
				result.WriteString(", ")
			}
		}
		result.WriteString(t)
	}
	return result.String()
}
