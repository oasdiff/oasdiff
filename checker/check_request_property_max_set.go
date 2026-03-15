package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxSetId              = "request-body-max-set"
	RequestPropertyMaxSetId          = "request-property-max-set"
	RequestBodyExclusiveMaxSetId     = "request-body-exclusive-max-set"
	RequestPropertyExclusiveMaxSetId = "request-property-exclusive-max-set"
)

func RequestPropertyMaxSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				_, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "maximum")
				if mediaTypeDiff.SchemaDiff.MaxDiff != nil {
					maxDiff := mediaTypeDiff.SchemaDiff.MaxDiff
					if maxDiff.From == nil &&
						maxDiff.To != nil {
						result = append(result, NewApiChange(
							RequestBodyMaxSetId,
							config,
							[]any{maxDiff.To},
							commentId(RequestBodyMaxSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
				}
				if mediaTypeDiff.SchemaDiff.ExclusiveMaxDiff != nil {
					exMaxDiff := mediaTypeDiff.SchemaDiff.ExclusiveMaxDiff
					if exMaxDiff.From == nil &&
						exMaxDiff.To != nil {
						_, exRevisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "exclusiveMaximum")
						result = append(result, NewApiChange(
							RequestBodyExclusiveMaxSetId,
							config,
							[]any{exMaxDiff.To},
							commentId(RequestBodyExclusiveMaxSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, exRevisionSource).WithDetails(mediaTypeDetails))
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						maxDiff := propertyDiff.MaxDiff
						if maxDiff == nil {
							return
						}
						if maxDiff.From != nil ||
							maxDiff.To == nil {
							return
						}
						if propertyDiff.Revision.ReadOnly {
							return
						}

						_, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "maximum")
						result = append(result, NewApiChange(
							RequestPropertyMaxSetId,
							config,
							[]any{propertyFullName(propertyPath, propertyName), maxDiff.To},
							commentId(RequestPropertyMaxSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
					})

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						exMaxDiff := propertyDiff.ExclusiveMaxDiff
						if exMaxDiff == nil {
							return
						}
						if exMaxDiff.From != nil ||
							exMaxDiff.To == nil {
							return
						}
						if propertyDiff.Revision.ReadOnly {
							return
						}

						_, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "exclusiveMaximum")
						result = append(result, NewApiChange(
							RequestPropertyExclusiveMaxSetId,
							config,
							[]any{propertyFullName(propertyPath, propertyName), exMaxDiff.To},
							commentId(RequestPropertyExclusiveMaxSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
					})
			}
		}
	}
	return result
}
