package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyIfAddedId       = "request-body-if-added"
	RequestBodyIfRemovedId     = "request-body-if-removed"
	RequestBodyThenAddedId     = "request-body-then-added"
	RequestBodyThenRemovedId   = "request-body-then-removed"
	RequestBodyElseAddedId     = "request-body-else-added"
	RequestBodyElseRemovedId   = "request-body-else-removed"
	RequestPropertyIfAddedId   = "request-property-if-added"
	RequestPropertyIfRemovedId = "request-property-if-removed"

	RequestPropertyThenAddedId   = "request-property-then-added"
	RequestPropertyThenRemovedId = "request-property-then-removed"
	RequestPropertyElseAddedId   = "request-property-else-added"
	RequestPropertyElseRemovedId = "request-property-else-removed"
)

func RequestPropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				for _, entry := range []struct {
					schemaDiff *diff.SchemaDiff
					addedId    string
					removedId  string
					field      string
				}{
					{mediaTypeDiff.SchemaDiff.IfDiff, RequestBodyIfAddedId, RequestBodyIfRemovedId, "if"},
					{mediaTypeDiff.SchemaDiff.ThenDiff, RequestBodyThenAddedId, RequestBodyThenRemovedId, "then"},
					{mediaTypeDiff.SchemaDiff.ElseDiff, RequestBodyElseAddedId, RequestBodyElseRemovedId, "else"},
				} {
					if entry.schemaDiff == nil {
						continue
					}
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, entry.field)
					if entry.schemaDiff.SchemaAdded {
						result = append(result, NewApiChange(
							entry.addedId,
							config,
							nil,
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
							nil,
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
							{propertyDiff.IfDiff, RequestPropertyIfAddedId, RequestPropertyIfRemovedId, "if"},
							{propertyDiff.ThenDiff, RequestPropertyThenAddedId, RequestPropertyThenRemovedId, "then"},
							{propertyDiff.ElseDiff, RequestPropertyElseAddedId, RequestPropertyElseRemovedId, "else"},
						} {
							if entry.schemaDiff == nil {
								continue
							}
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, entry.field)
							if entry.schemaDiff.SchemaAdded {
								result = append(result, NewApiChange(
									entry.addedId,
									config,
									[]any{propName},
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
									[]any{propName},
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
	return result
}
