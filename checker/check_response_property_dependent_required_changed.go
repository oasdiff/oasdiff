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
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.DependentRequiredDiff != nil {
						depReqDiff := mediaTypeDiff.SchemaDiff.DependentRequiredDiff
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "dependentRequired")
						fromMap, _ := depReqDiff.From.(map[string][]string)
						toMap, _ := depReqDiff.To.(map[string][]string)
						if len(fromMap) == 0 {
							result = append(result, NewApiChange(
								ResponseBodyDependentRequiredAddedId,
								config,
								[]any{depReqDiff.To, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						} else if len(toMap) == 0 {
							result = append(result, NewApiChange(
								ResponseBodyDependentRequiredRemovedId,
								config,
								[]any{depReqDiff.From, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								ResponseBodyDependentRequiredChangedId,
								config,
								[]any{depReqDiff.From, depReqDiff.To, responseStatus},
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
									ResponsePropertyDependentRequiredAddedId,
									config,
									[]any{propName, depReqDiff.To, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							} else if len(toMap) == 0 {
								result = append(result, NewApiChange(
									ResponsePropertyDependentRequiredRemovedId,
									config,
									[]any{propName, depReqDiff.From, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
							} else {
								result = append(result, NewApiChange(
									ResponsePropertyDependentRequiredChangedId,
									config,
									[]any{propName, depReqDiff.From, depReqDiff.To, responseStatus},
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
	}
	return result
}
