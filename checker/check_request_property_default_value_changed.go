package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyDefaultValueAddedId       = "request-body-default-value-added"
	RequestBodyDefaultValueRemovedId     = "request-body-default-value-removed"
	RequestBodyDefaultValueChangedId     = "request-body-default-value-changed"
	RequestPropertyDefaultValueAddedId   = "request-property-default-value-added"
	RequestPropertyDefaultValueRemovedId = "request-property-default-value-removed"
	RequestPropertyDefaultValueChangedId = "request-property-default-value-changed"
)

func RequestPropertyDefaultValueChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if mediaTypeDiff.SchemaDiff == nil {
					continue
				}
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "default")
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
					).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
				}
				if mediaTypeDiff.SchemaDiff.DefaultDiff != nil {
					defaultValueDiff := mediaTypeDiff.SchemaDiff.DefaultDiff

					if defaultValueDiff.From == nil {
						appendResultItem(RequestBodyDefaultValueAddedId, mediaType, defaultValueDiff.To)
					} else if defaultValueDiff.To == nil {
						appendResultItem(RequestBodyDefaultValueRemovedId, mediaType, defaultValueDiff.From)
					} else {
						appendResultItem(RequestBodyDefaultValueChangedId, mediaType, defaultValueDiff.From, defaultValueDiff.To)
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff == nil || propertyDiff.DefaultDiff == nil {
							return
						}

						defaultValueDiff := propertyDiff.DefaultDiff
						propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "default")

						appendPropResultItem := func(messageId string, a ...any) {
							result = append(result, NewApiChange(
								messageId,
								config,
								a,
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
						}

						if defaultValueDiff.From == nil {
							appendPropResultItem(RequestPropertyDefaultValueAddedId, propertyName, defaultValueDiff.To)
						} else if defaultValueDiff.To == nil {
							appendPropResultItem(RequestPropertyDefaultValueRemovedId, propertyName, defaultValueDiff.From)
						} else {
							appendPropResultItem(RequestPropertyDefaultValueChangedId, propertyName, defaultValueDiff.From, defaultValueDiff.To)
						}
					})
			}
		}
	}
	return result
}
