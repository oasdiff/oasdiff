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

							// if new schema is a subset of the old schema, it's not a breaking change
							if oneOfContains(propertyDiff) {
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

// oneOfContains checks if the new schema is a subset of oneOf in the old schema
func oneOfContains(propertyDiff *diff.SchemaDiff) bool {
	newOneOf := propertyDiff.Revision.OneOf
	// if the new schema isn't a oneOf, use the single schema
	if propertyDiff.Revision.OneOf == nil {
		newOneOf = []*openapi3.SchemaRef{openapi3.NewSchemaRef("", propertyDiff.Revision)}
	}

	// for each schema in the new oneOf, check if it exists in the old oneOf
	for _, newSchema := range newOneOf {
		if !findEquivalentSchema(newSchema, propertyDiff.Base.OneOf) {
			return false
		}
	}

	return true
}

func findEquivalentSchema(newSchema *openapi3.SchemaRef, oldOneOf []*openapi3.SchemaRef) bool {
	for _, oldSchema := range oldOneOf {
		if schemasEquivalent(newSchema, oldSchema) {
			return true
		}
	}
	return false
}

func schemasEquivalent(newSchema, oldSchema *openapi3.SchemaRef) bool {
	// if both schemas have a type only, compare the types
	if typeOnlySchema(newSchema) && typeOnlySchema(oldSchema) {
		return slices.Equal(newSchema.Value.Type.Slice(), oldSchema.Value.Type.Slice())
	}

	// TODO: identify other kinds of equivalence

	return false
}

func typeOnlySchema(schemaRef *openapi3.SchemaRef) bool {
	schema := *schemaRef.Value

	return schema.Extensions == nil &&
		schema.OneOf == nil &&
		schema.AnyOf == nil &&
		schema.AllOf == nil &&
		schema.Not == nil &&
		schema.Type != nil && // type is required
		schema.Title == "" &&
		schema.Format == "" &&
		schema.Description == "" &&
		schema.Enum == nil &&
		schema.Default == nil &&
		schema.Example == nil &&
		schema.ExternalDocs == nil &&
		!schema.UniqueItems &&
		!schema.ExclusiveMin &&
		!schema.ExclusiveMax &&
		!schema.Nullable &&
		!schema.ReadOnly &&
		!schema.WriteOnly &&
		!schema.AllowEmptyValue &&
		!schema.Deprecated &&
		schema.XML == nil &&
		schema.Min == nil &&
		schema.Max == nil &&
		schema.MultipleOf == nil &&
		schema.MinLength == 0 &&
		schema.MaxLength == nil &&
		schema.Pattern == "" &&
		schema.MinItems == 0 &&
		schema.MaxItems == nil &&
		schema.Items == nil &&
		schema.Required == nil &&
		schema.Properties == nil &&
		schema.MinProps == 0 &&
		schema.MaxProps == nil &&
		schema.AdditionalProperties == (openapi3.AdditionalProperties{
			Has:    nil,
			Schema: nil,
		}) &&
		schema.Discriminator == nil
}
