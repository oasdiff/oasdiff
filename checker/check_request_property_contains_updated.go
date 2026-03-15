package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyContainsAddedId            = "request-body-contains-added"
	RequestBodyContainsRemovedId          = "request-body-contains-removed"
	RequestBodyMinContainsIncreasedId     = "request-body-min-contains-increased"
	RequestBodyMinContainsDecreasedId     = "request-body-min-contains-decreased"
	RequestBodyMaxContainsIncreasedId     = "request-body-max-contains-increased"
	RequestBodyMaxContainsDecreasedId     = "request-body-max-contains-decreased"
	RequestPropertyContainsAddedId        = "request-property-contains-added"
	RequestPropertyContainsRemovedId      = "request-property-contains-removed"
	RequestPropertyMinContainsIncreasedId = "request-property-min-contains-increased"
	RequestPropertyMinContainsDecreasedId = "request-property-min-contains-decreased"
	RequestPropertyMaxContainsIncreasedId = "request-property-max-contains-increased"
	RequestPropertyMaxContainsDecreasedId = "request-property-max-contains-decreased"
)

func RequestPropertyContainsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				if mediaTypeDiff.SchemaDiff.ContainsDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "contains")
					if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaAdded {
						result = append(result, NewApiChange(
							RequestBodyContainsAddedId,
							config,
							nil,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
					if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaDeleted {
						result = append(result, NewApiChange(
							RequestBodyContainsRemovedId,
							config,
							nil,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
					}
				}

				if mediaTypeDiff.SchemaDiff.MinContainsDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "minContains")
					d := mediaTypeDiff.SchemaDiff.MinContainsDiff
					if IsIncreasedValue(d) {
						result = append(result, NewApiChange(
							RequestBodyMinContainsIncreasedId,
							config,
							[]any{d.From, d.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					}
					if IsDecreasedValue(d) {
						result = append(result, NewApiChange(
							RequestBodyMinContainsDecreasedId,
							config,
							[]any{d.From, d.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					}
				}

				if mediaTypeDiff.SchemaDiff.MaxContainsDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "maxContains")
					d := mediaTypeDiff.SchemaDiff.MaxContainsDiff
					if IsIncreasedValue(d) {
						result = append(result, NewApiChange(
							RequestBodyMaxContainsIncreasedId,
							config,
							[]any{d.From, d.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					}
					if IsDecreasedValue(d) {
						result = append(result, NewApiChange(
							RequestBodyMaxContainsDecreasedId,
							config,
							[]any{d.From, d.To},
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
						propName := propertyFullName(propertyPath, propertyName)

						if propertyDiff.ContainsDiff != nil {
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "contains")
							if propertyDiff.ContainsDiff.SchemaAdded {
								result = append(result, NewApiChange(
									RequestPropertyContainsAddedId,
									config,
									[]any{propName},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							if propertyDiff.ContainsDiff.SchemaDeleted {
								result = append(result, NewApiChange(
									RequestPropertyContainsRemovedId,
									config,
									[]any{propName},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
							}
						}

						if propertyDiff.MinContainsDiff != nil {
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "minContains")
							d := propertyDiff.MinContainsDiff
							if IsIncreasedValue(d) {
								result = append(result, NewApiChange(
									RequestPropertyMinContainsIncreasedId,
									config,
									[]any{propName, d.From, d.To},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							if IsDecreasedValue(d) {
								result = append(result, NewApiChange(
									RequestPropertyMinContainsDecreasedId,
									config,
									[]any{propName, d.From, d.To},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
							}
						}

						if propertyDiff.MaxContainsDiff != nil {
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "maxContains")
							d := propertyDiff.MaxContainsDiff
							if IsIncreasedValue(d) {
								result = append(result, NewApiChange(
									RequestPropertyMaxContainsIncreasedId,
									config,
									[]any{propName, d.From, d.To},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							if IsDecreasedValue(d) {
								result = append(result, NewApiChange(
									RequestPropertyMaxContainsDecreasedId,
									config,
									[]any{propName, d.From, d.To},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
							}
						}
					})
			}
		}
	}
	return result
}
