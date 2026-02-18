package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyDefaultValueAddedId       = "response-body-default-value-added"
	ResponseBodyDefaultValueRemovedId     = "response-body-default-value-removed"
	ResponseBodyDefaultValueChangedId     = "response-body-default-value-changed"
	ResponsePropertyDefaultValueAddedId   = "response-property-default-value-added"
	ResponsePropertyDefaultValueRemovedId = "response-property-default-value-removed"
	ResponsePropertyDefaultValueChangedId = "response-property-default-value-changed"
)

func ResponsePropertyDefaultValueChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			appendResultItem := func(messageId string, a ...any) {
				baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
				result = append(result, NewApiChange(
					messageId,
					config,
					a,
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
			}

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.DefaultDiff != nil {
						defaultValueDiff := mediaTypeDiff.SchemaDiff.DefaultDiff
						if defaultValueDiff.From == nil {
							appendResultItem(ResponseBodyDefaultValueAddedId, mediaType, defaultValueDiff.To, responseStatus)
						} else if defaultValueDiff.To == nil {
							appendResultItem(ResponseBodyDefaultValueRemovedId, mediaType, defaultValueDiff.From, responseStatus)
						} else {
							appendResultItem(ResponseBodyDefaultValueChangedId, mediaType, defaultValueDiff.From, defaultValueDiff.To, responseStatus)
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff == nil || propertyDiff.Revision == nil || propertyDiff.DefaultDiff == nil {
								return
							}

							defaultValueDiff := propertyDiff.DefaultDiff
							if defaultValueDiff.From == nil {
								appendResultItem(ResponsePropertyDefaultValueAddedId, propertyName, defaultValueDiff.To, responseStatus)
							} else if defaultValueDiff.To == nil {
								appendResultItem(ResponsePropertyDefaultValueRemovedId, propertyName, defaultValueDiff.From, responseStatus)
							} else {
								appendResultItem(ResponsePropertyDefaultValueChangedId, propertyName, defaultValueDiff.From, defaultValueDiff.To, responseStatus)
							}
						})
				}
			}
		}
	}
	return result
}
