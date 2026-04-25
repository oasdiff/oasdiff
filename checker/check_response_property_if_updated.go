package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyIfAddedId       = "response-body-if-added"
	ResponseBodyIfRemovedId     = "response-body-if-removed"
	ResponseBodyThenAddedId     = "response-body-then-added"
	ResponseBodyThenRemovedId   = "response-body-then-removed"
	ResponseBodyElseAddedId     = "response-body-else-added"
	ResponseBodyElseRemovedId   = "response-body-else-removed"
	ResponsePropertyIfAddedId   = "response-property-if-added"
	ResponsePropertyIfRemovedId = "response-property-if-removed"

	ResponsePropertyThenAddedId   = "response-property-then-added"
	ResponsePropertyThenRemovedId = "response-property-then-removed"
	ResponsePropertyElseAddedId   = "response-property-else-added"
	ResponsePropertyElseRemovedId = "response-property-else-removed"
)

func ResponsePropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					for _, entry := range []struct {
						schemaDiff *diff.SchemaDiff
						addedId    string
						removedId  string
						field      string
					}{
						{mediaTypeDiff.SchemaDiff.IfDiff, ResponseBodyIfAddedId, ResponseBodyIfRemovedId, "if"},
						{mediaTypeDiff.SchemaDiff.ThenDiff, ResponseBodyThenAddedId, ResponseBodyThenRemovedId, "then"},
						{mediaTypeDiff.SchemaDiff.ElseDiff, ResponseBodyElseAddedId, ResponseBodyElseRemovedId, "else"},
					} {
						if entry.schemaDiff == nil {
							continue
						}
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, entry.field)
						if entry.schemaDiff.SchemaAdded {
							result = append(result, NewApiChange(
								entry.addedId,
								config,
								[]any{responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						if entry.schemaDiff.SchemaDeleted {
							result = append(result, NewApiChange(
								entry.removedId,
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

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							propName := propertyFullName(propertyPath, propertyName)

							for _, entry := range []struct {
								schemaDiff *diff.SchemaDiff
								addedId    string
								removedId  string
								field      string
							}{
								{propertyDiff.IfDiff, ResponsePropertyIfAddedId, ResponsePropertyIfRemovedId, "if"},
								{propertyDiff.ThenDiff, ResponsePropertyThenAddedId, ResponsePropertyThenRemovedId, "then"},
								{propertyDiff.ElseDiff, ResponsePropertyElseAddedId, ResponsePropertyElseRemovedId, "else"},
							} {
								if entry.schemaDiff == nil {
									continue
								}
								propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, entry.field)
								if entry.schemaDiff.SchemaAdded {
									result = append(result, NewApiChange(
										entry.addedId,
										config,
										[]any{propName, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
								}
								if entry.schemaDiff.SchemaDeleted {
									result = append(result, NewApiChange(
										entry.removedId,
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
						})
				}
			}
		}
	}
	return result
}
