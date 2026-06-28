package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyOneOfAddedId       = "request-body-one-of-added"
	RequestBodyOneOfRemovedId     = "request-body-one-of-removed"
	RequestPropertyOneOfAddedId   = "request-property-one-of-added"
	RequestPropertyOneOfRemovedId = "request-property-one-of-removed"
)

func RequestPropertyOneOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// Body-level suppression also skips the property walk for this
		// media type (pre-migration code used `continue` in the
		// media-type for-loop). Preserved.
		if shouldSuppressOneOfSchemaChangedForListOfTypes(info.schemaDiff) {
			return
		}

		// A oneOf wrapping (#702) is reported once per body as
		// request-body-wrapped-in-one-of; don't also report its added
		// alternatives as a raw one-of-added.
		if !info.schemaDiff.OneOfWrappingDiff.Empty() {
			return
		}

		if info.schemaDiff.OneOfDiff != nil {
			if added := info.schemaDiff.OneOfDiff.Added; len(added) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "oneOf", -1, added[0].Index)
				result = append(result, info.newChange(RequestBodyOneOfAddedId, []any{added.String()}, "").
					WithSources(baseSource, revisionSource))
			}
			if deleted := info.schemaDiff.OneOfDiff.Deleted; len(deleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "oneOf", deleted[0].Index, -1)
				result = append(result, info.newChange(RequestBodyOneOfRemovedId, []any{deleted.String()}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.OneOfDiff == nil {
				return
			}
			if shouldSuppressPropertyOneOfSchemaChangedForListOfTypes(p.propertyDiff) {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if added := p.propertyDiff.OneOfDiff.Added; len(added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "oneOf", -1, added[0].Index)
				result = append(result, p.newChange(RequestPropertyOneOfAddedId, []any{added.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if deleted := p.propertyDiff.OneOfDiff.Deleted; len(deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "oneOf", deleted[0].Index, -1)
				result = append(result, p.newChange(RequestPropertyOneOfRemovedId, []any{deleted.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
