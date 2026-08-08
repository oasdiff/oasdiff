package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMediaTypeSchemaAddedId    = "request-body-media-type-schema-added"
	RequestBodyMediaTypeSchemaRemovedId  = "request-body-media-type-schema-removed"
	ResponseBodyMediaTypeSchemaAddedId   = "response-body-media-type-schema-added"
	ResponseBodyMediaTypeSchemaRemovedId = "response-body-media-type-schema-removed"

	RequestBodyMediaTypeItemSchemaAddedId    = "request-body-media-type-item-schema-added"
	RequestBodyMediaTypeItemSchemaRemovedId  = "request-body-media-type-item-schema-removed"
	ResponseBodyMediaTypeItemSchemaAddedId   = "response-body-media-type-item-schema-added"
	ResponseBodyMediaTypeItemSchemaRemovedId = "response-body-media-type-item-schema-removed"
)

// MediaTypeSchemaExistenceCheck classifies a schema added to, or removed from, a
// media type that exists on both sides (the media type itself is unchanged, only
// its schema appears or disappears). The per-schema-change checks skip these
// one-sided schema diffs (see modifiedSchemaPresentBothSides / #1047), so this
// check is what reports them, by request/response contravariance:
//
//	request body, schema added   -> breaking (ERR): a body that accepted anything
//	                                now requires a specific shape (narrowing)
//	request body, schema removed -> info: the body got more permissive
//	response, schema added       -> info: the response got more specific
//	response, schema removed     -> breaking (WARN): a typed response is no longer
//	                                guaranteed (the output contract loosened)
//
// The OpenAPI 3.2 itemSchema, which types one item of a streamed body, is
// classified the same way and for the same reasons: it constrains what may be
// sent or returned, so adding one narrows and removing one loosens.
func MediaTypeSchemaExistenceCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)

			if operationItem.RequestBodyDiff != nil && operationItem.RequestBodyDiff.ContentDiff != nil {
				for mediaType, mediaTypeDiff := range operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified {
					for _, schema := range requestSchemaKinds(mediaTypeDiff) {
						added, removed := schemaSideAddedRemoved(schema.diff)
						if added {
							result = append(result, opInfo.NewApiChange(
								schema.addedId, []any{mediaType}, "",
							).WithSources(nil, requestBodyMediaTypeSource(operationsSources, operationItem.Revision, mediaType)))
						} else if removed {
							result = append(result, opInfo.NewApiChange(
								schema.removedId, []any{mediaType}, "",
							).WithSources(requestBodyMediaTypeSource(operationsSources, operationItem.Base, mediaType), nil))
						}
					}
				}
			}

			if operationItem.ResponsesDiff == nil {
				continue
			}
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil {
					continue
				}
				for mediaType, mediaTypeDiff := range responseDiff.ContentDiff.MediaTypeModified {
					for _, schema := range responseSchemaKinds(mediaTypeDiff) {
						added, removed := schemaSideAddedRemoved(schema.diff)
						if added {
							result = append(result, opInfo.NewApiChange(
								schema.addedId, []any{mediaType, responseStatus}, "",
							).WithSources(nil, mediaTypeSource(operationsSources, operationItem.Revision, responseDiff.Revision, mediaType)))
						} else if removed {
							result = append(result, opInfo.NewApiChange(
								schema.removedId, []any{mediaType, responseStatus}, "",
							).WithSources(mediaTypeSource(operationsSources, operationItem.Base, responseDiff.Base, mediaType), nil))
						}
					}
				}
			}
		}
	}

	return result
}

// schemaKind pairs one of a media type's schemas with the check IDs that report
// it appearing or disappearing.
type schemaKind struct {
	diff      *diff.SchemaDiff
	addedId   string
	removedId string
}

func requestSchemaKinds(d *diff.MediaTypeDiff) []schemaKind {
	return []schemaKind{
		{d.SchemaDiff, RequestBodyMediaTypeSchemaAddedId, RequestBodyMediaTypeSchemaRemovedId},
		{d.ItemSchemaDiff, RequestBodyMediaTypeItemSchemaAddedId, RequestBodyMediaTypeItemSchemaRemovedId},
	}
}

func responseSchemaKinds(d *diff.MediaTypeDiff) []schemaKind {
	return []schemaKind{
		{d.SchemaDiff, ResponseBodyMediaTypeSchemaAddedId, ResponseBodyMediaTypeSchemaRemovedId},
		{d.ItemSchemaDiff, ResponseBodyMediaTypeItemSchemaAddedId, ResponseBodyMediaTypeItemSchemaRemovedId},
	}
}

// schemaSideAddedRemoved reports whether a media type's schema diff represents a
// schema added or removed. The diff sets the explicit SchemaAdded / SchemaDeleted
// flags for these one-sided cases (Base and Revision are both nil then). Returns
// false, false when the schema changed on both sides or is absent.
func schemaSideAddedRemoved(d *diff.SchemaDiff) (added, removed bool) {
	if d == nil {
		return false, false
	}
	return d.SchemaAdded, d.SchemaDeleted
}
