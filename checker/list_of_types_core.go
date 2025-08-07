package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// Core logic functions for ListOfTypes checking that can be reused by both full checkers and suppression functions

// checkPropertyListOfTypesChange checks if a property change involves list-of-types patterns and returns breaking changes
func checkPropertyListOfTypesChange(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff,
	mediaType string, responseStatus string, config *Config, operationsSources *diff.OperationsSourcesMap,
	operationItem *diff.MethodDiff, operation string, path string, isRequest bool) Changes {

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

	result = append(result, NewApiChange(
		messageId,
		config,
		args,
		"",
		operationsSources,
		operationItem.Revision,
		operation,
		path,
	))

	return result
}

// checkBodyListOfTypesChange checks if a request/response body change involves list-of-types patterns and returns breaking changes
func checkBodyListOfTypesChange(schemaDiff *diff.SchemaDiff, mediaType string, responseStatus string,
	config *Config, operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff,
	operation string, path string, isRequest bool) Changes {

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

	result = append(result, NewApiChange(
		messageId,
		config,
		args,
		"",
		operationsSources,
		operationItem.Revision,
		operation,
		path,
	))

	return result
}

// checkParameterListOfTypesChange checks if a parameter change involves list-of-types patterns and returns breaking changes
func checkParameterListOfTypesChange(paramDiff *diff.ParameterDiff, param *openapi3.Parameter,
	config *Config, operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff,
	operation string, path string) Changes {

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

	result = append(result, NewApiChange(
		messageId,
		config,
		args,
		"",
		operationsSources,
		operationItem.Revision,
		operation,
		path,
	))

	return result
}

// checkParameterPropertyListOfTypesChange checks if a parameter property change involves list-of-types patterns and returns breaking changes
func checkParameterPropertyListOfTypesChange(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff,
	param *openapi3.Parameter, config *Config, operationsSources *diff.OperationsSourcesMap,
	operationItem *diff.MethodDiff, operation string, path string) Changes {

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

	result = append(result, NewApiChange(
		messageId,
		config,
		args,
		"",
		operationsSources,
		operationItem.Revision,
		operation,
		path,
	))

	return result
}

// Suppression functions that can be called by existing checkers

// shouldSuppressTypeChangedForListOfTypes checks if a type change should be suppressed because it's handled by ListOfTypes checker
func shouldSuppressTypeChangedForListOfTypes(schemaDiff *diff.SchemaDiff) bool {
	return schemaDiff != nil && !schemaDiff.ListOfTypesDiff.Empty()
}

// shouldSuppressPropertyTypeChangedForListOfTypes checks if a property type change should be suppressed because it's handled by ListOfTypes checker
func shouldSuppressPropertyTypeChangedForListOfTypes(propertyDiff *diff.SchemaDiff) bool {
	return propertyDiff != nil && !propertyDiff.ListOfTypesDiff.Empty()
}

// shouldSuppressOneOfSchemaChangedForListOfTypes checks if oneOf schema changes should be suppressed because they're handled by ListOfTypes checker
func shouldSuppressOneOfSchemaChangedForListOfTypes(schemaDiff *diff.SchemaDiff) bool {
	return schemaDiff != nil && !schemaDiff.ListOfTypesDiff.Empty()
}

// shouldSuppressPropertyOneOfSchemaChangedForListOfTypes checks if property oneOf schema changes should be suppressed because they're handled by ListOfTypes checker
func shouldSuppressPropertyOneOfSchemaChangedForListOfTypes(propertyDiff *diff.SchemaDiff) bool {
	return propertyDiff != nil && !propertyDiff.ListOfTypesDiff.Empty()
}

// Helper function to join types for display in messages
func joinTypes(types []string) string {
	if len(types) == 0 {
		return ""
	}
	if len(types) == 1 {
		return types[0]
	}

	result := ""
	for i, t := range types {
		if i > 0 {
			if i == len(types)-1 {
				result += " and "
			} else {
				result += ", "
			}
		}
		result += t
	}
	return result
}
