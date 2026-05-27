package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	RequestBodyAnyOfAddedId       = "request-body-any-of-added"
	RequestBodyAnyOfRemovedId     = "request-body-any-of-removed"
	RequestPropertyAnyOfAddedId   = "request-property-any-of-added"
	RequestPropertyAnyOfRemovedId = "request-property-any-of-removed"
)

func RequestPropertyAnyOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// Body-level suppression also skips the property walk for this
		// media type (pre-migration code used `continue` in the
		// media-type for-loop). Preserved.
		if shouldSuppressOneOfSchemaChangedForListOfTypes(info.schemaDiff) {
			return
		}

		if info.schemaDiff.AnyOfDiff != nil {
			if added := info.schemaDiff.AnyOfDiff.Added; len(added) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "anyOf", -1, added[0].Index)
				result = append(result, info.newChange(RequestBodyAnyOfAddedId, []any{added.String()}, "").
					WithSources(baseSource, revisionSource))
			}
			if deleted := info.schemaDiff.AnyOfDiff.Deleted; len(deleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "anyOf", deleted[0].Index, -1)
				result = append(result, info.newChange(RequestBodyAnyOfRemovedId, []any{deleted.String()}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.AnyOfDiff == nil {
				return
			}
			if shouldSuppressPropertyOneOfSchemaChangedForListOfTypes(p.propertyDiff) {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if added := p.propertyDiff.AnyOfDiff.Added; len(added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "anyOf", -1, added[0].Index)
				result = append(result, p.newChange(RequestPropertyAnyOfAddedId, []any{added.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if deleted := p.propertyDiff.AnyOfDiff.Deleted; len(deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "anyOf", deleted[0].Index, -1)
				result = append(result, p.newChange(RequestPropertyAnyOfRemovedId, []any{deleted.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
