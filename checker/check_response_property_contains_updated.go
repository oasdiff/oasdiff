package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyContainsAddedId            = "response-body-contains-added"
	ResponseBodyContainsRemovedId          = "response-body-contains-removed"
	ResponseBodyMinContainsIncreasedId     = "response-body-min-contains-increased"
	ResponseBodyMinContainsDecreasedId     = "response-body-min-contains-decreased"
	ResponseBodyMaxContainsIncreasedId     = "response-body-max-contains-increased"
	ResponseBodyMaxContainsDecreasedId     = "response-body-max-contains-decreased"
	ResponsePropertyContainsAddedId        = "response-property-contains-added"
	ResponsePropertyContainsRemovedId      = "response-property-contains-removed"
	ResponsePropertyMinContainsIncreasedId = "response-property-min-contains-increased"
	ResponsePropertyMinContainsDecreasedId = "response-property-min-contains-decreased"
	ResponsePropertyMaxContainsIncreasedId = "response-property-max-contains-increased"
	ResponsePropertyMaxContainsDecreasedId = "response-property-max-contains-decreased"
)

func ResponsePropertyContainsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					if mediaTypeDiff.SchemaDiff.ContainsDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "contains")
						if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaAdded {
							result = append(result, NewApiChange(
								ResponseBodyContainsAddedId,
								config,
								[]any{responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaDeleted {
							result = append(result, NewApiChange(
								ResponseBodyContainsRemovedId,
								config,
								[]any{responseStatus},
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
								ResponseBodyMinContainsIncreasedId,
								config,
								[]any{d.From, d.To, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						}
						if IsDecreasedValue(d) {
							result = append(result, NewApiChange(
								ResponseBodyMinContainsDecreasedId,
								config,
								[]any{d.From, d.To, responseStatus},
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
								ResponseBodyMaxContainsIncreasedId,
								config,
								[]any{d.From, d.To, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						}
						if IsDecreasedValue(d) {
							result = append(result, NewApiChange(
								ResponseBodyMaxContainsDecreasedId,
								config,
								[]any{d.From, d.To, responseStatus},
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
										ResponsePropertyContainsAddedId,
										config,
										[]any{propName, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
								}
								if propertyDiff.ContainsDiff.SchemaDeleted {
									result = append(result, NewApiChange(
										ResponsePropertyContainsRemovedId,
										config,
										[]any{propName, responseStatus},
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
										ResponsePropertyMinContainsIncreasedId,
										config,
										[]any{propName, d.From, d.To, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
								}
								if IsDecreasedValue(d) {
									result = append(result, NewApiChange(
										ResponsePropertyMinContainsDecreasedId,
										config,
										[]any{propName, d.From, d.To, responseStatus},
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
										ResponsePropertyMaxContainsIncreasedId,
										config,
										[]any{propName, d.From, d.To, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
								}
								if IsDecreasedValue(d) {
									result = append(result, NewApiChange(
										ResponsePropertyMaxContainsDecreasedId,
										config,
										[]any{propName, d.From, d.To, responseStatus},
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
	}
	return result
}
