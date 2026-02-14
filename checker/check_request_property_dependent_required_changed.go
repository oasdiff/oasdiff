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
				appendResultItem := func(messageId string, a ...any) {
					result = append(result, NewApiChange(
						messageId,
						config,
						a,
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithDetails(mediaTypeDetails))
				}
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.DependentRequiredDiff != nil {
					depReqDiff := mediaTypeDiff.SchemaDiff.DependentRequiredDiff
					fromMap, _ := depReqDiff.From.(map[string][]string)
					toMap, _ := depReqDiff.To.(map[string][]string)
					if len(fromMap) == 0 {
						appendResultItem(RequestBodyDependentRequiredAddedId, depReqDiff.To)
					} else if len(toMap) == 0 {
						appendResultItem(RequestBodyDependentRequiredRemovedId, depReqDiff.From)
					} else {
						appendResultItem(RequestBodyDependentRequiredChangedId, depReqDiff.From, depReqDiff.To)
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
						fromMap, _ := depReqDiff.From.(map[string][]string)
						toMap, _ := depReqDiff.To.(map[string][]string)
						if len(fromMap) == 0 {
							appendResultItem(RequestPropertyDependentRequiredAddedId, propName, depReqDiff.To)
						} else if len(toMap) == 0 {
							appendResultItem(RequestPropertyDependentRequiredRemovedId, propName, depReqDiff.From)
						} else {
							appendResultItem(RequestPropertyDependentRequiredChangedId, propName, depReqDiff.From, depReqDiff.To)
						}
					})
			}
		}
	}
	return result
}
