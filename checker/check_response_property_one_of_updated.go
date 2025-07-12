package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyOneOfAddedId           = "response-body-one-of-added"
	ResponseBodyOneOfRemovedId         = "response-body-one-of-removed"
	ResponsePropertyOneOfAddedId       = "response-property-one-of-added"
	ResponsePropertyOneOfRemovedId     = "response-property-one-of-removed"
	ResponsePropertyOneOfSpecializedId = "response-property-one-of-specialized"
)

func ResponsePropertyOneOfUpdated(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			for responseStatus, responsesDiff := range operationItem.ResponsesDiff.Modified {
				if responsesDiff.ContentDiff == nil || responsesDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responsesDiff.ContentDiff.MediaTypeModified
				for _, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					// TODO: handle the case where the new schema is a subset of the old schema

					if mediaTypeDiff.SchemaDiff.OneOfDiff != nil && len(mediaTypeDiff.SchemaDiff.OneOfDiff.Added) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyOneOfAddedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.OneOfDiff.Added.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					if mediaTypeDiff.SchemaDiff.OneOfDiff != nil && len(mediaTypeDiff.SchemaDiff.OneOfDiff.Deleted) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyOneOfRemovedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.OneOfDiff.Deleted.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff.OneOfDiff == nil {
								return
							}

							propName := propertyFullName(propertyPath, propertyName)

							if typeSpecialized(propertyDiff) {
								result = append(result, NewApiChange(
									ResponsePropertyOneOfSpecializedId,
									config,
									[]any{propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								))
								return
							}

							if len(propertyDiff.OneOfDiff.Added) > 0 {
								result = append(result, NewApiChange(
									ResponsePropertyOneOfAddedId,
									config,
									[]any{propertyDiff.OneOfDiff.Added.String(), propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								))
							}

							if len(propertyDiff.OneOfDiff.Deleted) > 0 {
								result = append(result, NewApiChange(
									ResponsePropertyOneOfRemovedId,
									config,
									[]any{propertyDiff.OneOfDiff.Deleted.String(), propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								))
							}
						})
				}
			}
		}
	}
	return result
}

// typeSpecialized checks for a specific use case of oneOf:
// the original schema has a list of types under oneOf, and the new schema has a single type that is a subset of the original list
func typeSpecialized(propertyDiff *diff.SchemaDiff) bool {
	
	// check that all base schamas have a type only
	for _, schema := range propertyDiff.Base.OneOf {
		if !typeOnlySchema(schema.Value) {
			return false
		}
	}

	// check if revision type exists in the base oneOf
	if !typeOnlySchema(propertyDiff.Revision) {
		return false
	}

	for _, baseSchema := range propertyDiff.Base.OneOf {
		if slices.Equal(baseSchema.Value.Type.Slice(), propertyDiff.Revision.Type.Slice()) {
			return true
		}
	}
	return false
}

// typeOnlySchema checks if the schema has a type definition and no other properties
func typeOnlySchema(schema *openapi3.Schema) bool {
	copy := *schema
	copy.Type = nil
	return copy.IsEmpty()
}
