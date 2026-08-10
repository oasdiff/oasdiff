package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	// Regular schema
	RequestBodyMediaTypeSchemaAddedId    = "request-body-media-type-schema-added"
	RequestBodyMediaTypeSchemaRemovedId  = "request-body-media-type-schema-removed"
	ResponseBodyMediaTypeSchemaAddedId   = "response-body-media-type-schema-added"
	ResponseBodyMediaTypeSchemaRemovedId = "response-body-media-type-schema-removed"

	// OpenAPI 3.2 itemSchema
	RequestBodyMediaTypeItemSchemaAddedId           = "request-body-media-type-item-schema-added"
	RequestBodyMediaTypeItemSchemaRemovedId         = "request-body-media-type-item-schema-removed"
	ResponseBodyMediaTypeItemSchemaAddedId          = "response-body-media-type-item-schema-added"
	ResponseBodyMediaTypeItemSchemaRemovedId        = "response-body-media-type-item-schema-removed"
	ResponseBodyMediaTypeItemSchemaRemovedUntypedId = "response-body-media-type-item-schema-removed-untyped"
)

// MediaTypeSchemaExistenceCheck classifies a schema added to, or removed from, a
// media type that exists on both sides (the media type itself is unchanged, only
// its schema appears or disappears). The per-schema-change checks skip these
// one-sided schema diffs (see modifiedSchemaPresentBothSides), so this
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
					addedSource := mediaTypeSource(operationsSources, operationItem.Revision, responseDiff.Revision, mediaType)
					removedSource := mediaTypeSource(operationsSources, operationItem.Base, responseDiff.Base, mediaType)

					// Regular schema
					if added, removed := schemaSideAddedRemoved(mediaTypeDiff.SchemaDiff); added {
						result = append(result, opInfo.NewApiChange(
							ResponseBodyMediaTypeSchemaAddedId, []any{mediaType, responseStatus}, "",
						).WithSources(nil, addedSource))
					} else if removed {
						result = append(result, opInfo.NewApiChange(
							ResponseBodyMediaTypeSchemaRemovedId, []any{mediaType, responseStatus}, "",
						).WithSources(removedSource, nil))
					}

					// OpenAPI 3.2 itemSchema
					if added, removed := schemaSideAddedRemoved(mediaTypeDiff.ItemSchemaDiff); added {
						result = append(result, opInfo.NewApiChange(
							ResponseBodyMediaTypeItemSchemaAddedId, []any{mediaType, responseStatus}, "",
						).WithSources(nil, addedSource))
					} else if removed {
						// A remaining whole-body schema does not soften this. It and itemSchema
						// constrain different things, so `type: object` over a stream says nothing
						// about one item, and dropping the item schema widens what the server may
						// send either way. The id is separate only so a team that has decided this
						// case is safe for their ecosystem can downgrade it alone.
						id := ResponseBodyMediaTypeItemSchemaRemovedId
						if !responseMediaTypeHasSchema(responseDiff.Revision, mediaType) {
							// Nothing types the body at all now, not just its items.
							id = ResponseBodyMediaTypeItemSchemaRemovedUntypedId
						}
						result = append(result, opInfo.NewApiChange(
							id, []any{mediaType, responseStatus}, "",
						).WithSources(removedSource, nil))
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

// responseMediaTypeHasSchema reports whether a response media type declares a
// whole-body schema, which decides whether removing its item schema leaves the
// body typed.
func responseMediaTypeHasSchema(response *openapi3.Response, mediaType string) bool {
	if response == nil || response.Content == nil {
		return false
	}
	content := response.Content[mediaType]
	return content != nil && content.Schema != nil
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
