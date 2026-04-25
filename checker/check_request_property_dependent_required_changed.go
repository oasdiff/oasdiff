package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyDependentRequiredAddedId       = "request-body-dependent-required-added"
	RequestBodyDependentRequiredRemovedId     = "request-body-dependent-required-removed"
	RequestBodyDependentRequiredChangedId     = "request-body-dependent-required-changed"
	RequestPropertyDependentRequiredAddedId   = "request-property-dependent-required-added"
	RequestPropertyDependentRequiredRemovedId = "request-property-dependent-required-removed"
	RequestPropertyDependentRequiredChangedId = "request-property-dependent-required-changed"
)

func RequestPropertyDependentRequiredChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.DependentRequiredDiff != nil {
					depReqDiff := mediaTypeDiff.SchemaDiff.DependentRequiredDiff
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "dependentRequired")
					result = append(result, requestBodyDependentRequiredChanges(
						depReqDiff, config, operationsSources, operationItem, operation, path,
						baseSource, revisionSource, mediaTypeDetails,
					)...)
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff == nil || propertyDiff.DependentRequiredDiff == nil {
							return
						}

						depReqDiff := propertyDiff.DependentRequiredDiff
						propName := propertyFullName(propertyPath, propertyName)
						propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "dependentRequired")
						result = append(result, requestPropertyDependentRequiredChanges(
							depReqDiff, config, operationsSources, operationItem, operation, path,
							propName, propBaseSource, propRevisionSource, mediaTypeDetails,
						)...)
					})
			}
		}
	}
	return result
}

func requestBodyDependentRequiredChanges(
	depReqDiff *diff.DependentRequiredDiff,
	config *Config,
	operationsSources *diff.OperationsSourcesMap,
	operationItem *diff.MethodDiff,
	operation string,
	path string,
	baseSource *Source,
	revisionSource *Source,
	mediaTypeDetails string,
) Changes {
	result := make(Changes, 0)

	for key, values := range depReqDiff.Added {
		result = append(result, NewApiChange(
			RequestBodyDependentRequiredAddedId,
			config,
			[]any{key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
	}

	for key, values := range depReqDiff.Deleted {
		result = append(result, NewApiChange(
			RequestBodyDependentRequiredRemovedId,
			config,
			[]any{key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
	}

	for key, stringsDiff := range depReqDiff.Modified {
		result = append(result, NewApiChange(
			RequestBodyDependentRequiredChangedId,
			config,
			[]any{key, formatDependentRequiredModification(stringsDiff)},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
	}

	return result
}

func requestPropertyDependentRequiredChanges(
	depReqDiff *diff.DependentRequiredDiff,
	config *Config,
	operationsSources *diff.OperationsSourcesMap,
	operationItem *diff.MethodDiff,
	operation string,
	path string,
	propName string,
	baseSource *Source,
	revisionSource *Source,
	mediaTypeDetails string,
) Changes {
	result := make(Changes, 0)

	for key, values := range depReqDiff.Added {
		result = append(result, NewApiChange(
			RequestPropertyDependentRequiredAddedId,
			config,
			[]any{propName, key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
	}

	for key, values := range depReqDiff.Deleted {
		result = append(result, NewApiChange(
			RequestPropertyDependentRequiredRemovedId,
			config,
			[]any{propName, key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
	}

	for key, stringsDiff := range depReqDiff.Modified {
		result = append(result, NewApiChange(
			RequestPropertyDependentRequiredChangedId,
			config,
			[]any{propName, key, formatDependentRequiredModification(stringsDiff)},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
	}

	return result
}
