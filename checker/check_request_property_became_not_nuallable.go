package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyBecomeNotNullableId     = "request-body-became-not-nullable"
	RequestBodyBecomeNullableId        = "request-body-became-nullable"
	RequestPropertyBecomeNotNullableId = "request-property-became-not-nullable"
	RequestPropertyBecomeNullableId    = "request-property-became-nullable"
)

func RequestPropertyBecameNotNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if mediaTypeDiff.SchemaDiff == nil {
					continue
				}
				baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "nullable")

				if mediaTypeDiff.SchemaDiff.NullableDiff != nil {
					if mediaTypeDiff.SchemaDiff.NullableDiff.From == true {
						result = append(result, NewApiChange(
							RequestBodyBecomeNotNullableId,
							config,
							nil,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					} else if mediaTypeDiff.SchemaDiff.NullableDiff.To == true {
						result = append(result, NewApiChange(
							RequestBodyBecomeNullableId,
							config,
							nil,
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
						nullableDiff := propertyDiff.NullableDiff
						if nullableDiff == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)
						propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "nullable")

						if nullableDiff.From == true {
							result = append(result, NewApiChange(
								RequestPropertyBecomeNotNullableId,
								config,
								[]any{propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
						} else if nullableDiff.To == true {
							result = append(result, NewApiChange(
								RequestPropertyBecomeNullableId,
								config,
								[]any{propName},
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
