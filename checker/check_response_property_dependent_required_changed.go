package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyDependentRequiredAddedId       = "response-body-dependent-required-added"
	ResponseBodyDependentRequiredRemovedId     = "response-body-dependent-required-removed"
	ResponseBodyDependentRequiredChangedId     = "response-body-dependent-required-changed"
	ResponsePropertyDependentRequiredAddedId   = "response-property-dependent-required-added"
	ResponsePropertyDependentRequiredRemovedId = "response-property-dependent-required-removed"
	ResponsePropertyDependentRequiredChangedId = "response-property-dependent-required-changed"
)

func ResponsePropertyDependentRequiredChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil || responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.DependentRequiredDiff != nil {
						depReqDiff := mediaTypeDiff.SchemaDiff.DependentRequiredDiff
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "dependentRequired")
						result = append(result, responseBodyDependentRequiredChanges(
							depReqDiff, config, operationsSources, operationItem, operation, path,
							responseStatus, baseSource, revisionSource, mediaTypeDetails,
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
							result = append(result, responsePropertyDependentRequiredChanges(
								depReqDiff, config, operationsSources, operationItem, operation, path,
								responseStatus, propName, propBaseSource, propRevisionSource, mediaTypeDetails,
							)...)
						})
				}
			}
		}
	}
	return result
}

func responseBodyDependentRequiredChanges(
	depReqDiff *diff.DependentRequiredDiff,
	config *Config,
	operationsSources *diff.OperationsSourcesMap,
	operationItem *diff.MethodDiff,
	operation string,
	path string,
	responseStatus string,
	baseSource *Source,
	revisionSource *Source,
	mediaTypeDetails string,
) Changes {
	result := make(Changes, 0)

	// message: "the response body dependentRequired was added for the status %s: when '%s' is present, '%s' are required"
	for key, values := range depReqDiff.Added {
		result = append(result, NewApiChange(
			ResponseBodyDependentRequiredAddedId,
			config,
			[]any{responseStatus, key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
	}

	// message: "the response body dependentRequired was removed for the status %s: when '%s' was present, '%s' were required"
	for key, values := range depReqDiff.Deleted {
		result = append(result, NewApiChange(
			ResponseBodyDependentRequiredRemovedId,
			config,
			[]any{responseStatus, key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
	}

	// message: "the response body dependentRequired for '%s' was updated for the status %s: %s"
	for key, stringsDiff := range depReqDiff.Modified {
		result = append(result, NewApiChange(
			ResponseBodyDependentRequiredChangedId,
			config,
			[]any{key, responseStatus, formatDependentRequiredModification(stringsDiff)},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
	}

	return result
}

func responsePropertyDependentRequiredChanges(
	depReqDiff *diff.DependentRequiredDiff,
	config *Config,
	operationsSources *diff.OperationsSourcesMap,
	operationItem *diff.MethodDiff,
	operation string,
	path string,
	responseStatus string,
	propName string,
	baseSource *Source,
	revisionSource *Source,
	mediaTypeDetails string,
) Changes {
	result := make(Changes, 0)

	// message: "the %s response property dependentRequired was added for the status %s: when '%s' is present, '%s' are required"
	for key, values := range depReqDiff.Added {
		result = append(result, NewApiChange(
			ResponsePropertyDependentRequiredAddedId,
			config,
			[]any{propName, responseStatus, key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
	}

	// message: "the %s response property dependentRequired was removed for the status %s: when '%s' was present, '%s' were required"
	for key, values := range depReqDiff.Deleted {
		result = append(result, NewApiChange(
			ResponsePropertyDependentRequiredRemovedId,
			config,
			[]any{propName, responseStatus, key, strings.Join(values, ", ")},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
	}

	// message: "the %s response property dependentRequired for '%s' was updated for the status %s: %s"
	for key, stringsDiff := range depReqDiff.Modified {
		result = append(result, NewApiChange(
			ResponsePropertyDependentRequiredChangedId,
			config,
			[]any{propName, key, responseStatus, formatDependentRequiredModification(stringsDiff)},
			"",
			operationsSources,
			operationItem.Revision,
			operation,
			path,
		).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
	}

	return result
}
