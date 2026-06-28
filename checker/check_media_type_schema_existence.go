package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMediaTypeSchemaAddedId    = "request-body-media-type-schema-added"
	RequestBodyMediaTypeSchemaRemovedId  = "request-body-media-type-schema-removed"
	ResponseBodyMediaTypeSchemaAddedId   = "response-body-media-type-schema-added"
	ResponseBodyMediaTypeSchemaRemovedId = "response-body-media-type-schema-removed"
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

			if operationItem.RequestBodyDiff != nil && operationItem.RequestBodyDiff.ContentDiff != nil {
				for mediaType, mediaTypeDiff := range operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified {
					added, removed := schemaSideAddedRemoved(mediaTypeDiff.SchemaDiff)
					if added {
						result = append(result, NewApiChange(
							RequestBodyMediaTypeSchemaAddedId, config, []any{mediaType}, "",
							operationsSources, operationItem.Revision, operation, path,
						).WithSources(nil, requestBodyMediaTypeSource(operationsSources, operationItem.Revision, mediaType)))
					} else if removed {
						result = append(result, NewApiChange(
							RequestBodyMediaTypeSchemaRemovedId, config, []any{mediaType}, "",
							operationsSources, operationItem.Revision, operation, path,
						).WithSources(requestBodyMediaTypeSource(operationsSources, operationItem.Base, mediaType), nil))
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
					added, removed := schemaSideAddedRemoved(mediaTypeDiff.SchemaDiff)
					if added {
						result = append(result, NewApiChange(
							ResponseBodyMediaTypeSchemaAddedId, config, []any{mediaType, responseStatus}, "",
							operationsSources, operationItem.Revision, operation, path,
						).WithSources(nil, mediaTypeSource(operationsSources, operationItem.Revision, responseDiff.Revision, mediaType)))
					} else if removed {
						result = append(result, NewApiChange(
							ResponseBodyMediaTypeSchemaRemovedId, config, []any{mediaType, responseStatus}, "",
							operationsSources, operationItem.Revision, operation, path,
						).WithSources(mediaTypeSource(operationsSources, operationItem.Base, responseDiff.Base, mediaType), nil))
					}
				}
			}
		}
	}

	return result
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
