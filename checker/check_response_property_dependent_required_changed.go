package checker

import (
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
							appendResultItem(ResponseBodyDependentRequiredAddedId, depReqDiff.To, responseStatus)
						} else if len(toMap) == 0 {
							appendResultItem(ResponseBodyDependentRequiredRemovedId, depReqDiff.From, responseStatus)
						} else {
							appendResultItem(ResponseBodyDependentRequiredChangedId, depReqDiff.From, depReqDiff.To, responseStatus)
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
								appendResultItem(ResponsePropertyDependentRequiredAddedId, propName, depReqDiff.To, responseStatus)
							} else if len(toMap) == 0 {
								appendResultItem(ResponsePropertyDependentRequiredRemovedId, propName, depReqDiff.From, responseStatus)
							} else {
								appendResultItem(ResponsePropertyDependentRequiredChangedId, propName, depReqDiff.From, depReqDiff.To, responseStatus)
							}
						})
				}
			}
		}
	}
	return result
}
