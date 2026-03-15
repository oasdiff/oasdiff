package checker

import (
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
					fromMap, _ := depReqDiff.From.(map[string][]string)
					toMap, _ := depReqDiff.To.(map[string][]string)
					if len(fromMap) == 0 {
						result = append(result, NewApiChange(
							RequestBodyDependentRequiredAddedId,
							config,
							[]any{depReqDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					} else if len(toMap) == 0 {
						result = append(result, NewApiChange(
							RequestBodyDependentRequiredRemovedId,
							config,
							[]any{depReqDiff.From},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
					} else {
						result = append(result, NewApiChange(
							RequestBodyDependentRequiredChangedId,
							config,
							[]any{depReqDiff.From, depReqDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					}
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
						fromMap, _ := depReqDiff.From.(map[string][]string)
						toMap, _ := depReqDiff.To.(map[string][]string)
						if len(fromMap) == 0 {
							result = append(result, NewApiChange(
								RequestPropertyDependentRequiredAddedId,
								config,
								[]any{propName, depReqDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
						} else if len(toMap) == 0 {
							result = append(result, NewApiChange(
								RequestPropertyDependentRequiredRemovedId,
								config,
								[]any{propName, depReqDiff.From},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestPropertyDependentRequiredChangedId,
								config,
								[]any{propName, depReqDiff.From, depReqDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
						}
					})
			}
		}
	}
	return result
}
