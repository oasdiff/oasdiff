package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyDependentSchemaAddedId       = "request-body-dependent-schema-added"
	RequestBodyDependentSchemaRemovedId     = "request-body-dependent-schema-removed"
	RequestPropertyDependentSchemaAddedId   = "request-property-dependent-schema-added"
	RequestPropertyDependentSchemaRemovedId = "request-property-dependent-schema-removed"
)

func RequestPropertyDependentSchemasUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.DependentSchemasDiff != nil {
			depSchemasDiff := info.schemaDiff.DependentSchemasDiff
			for _, name := range depSchemasDiff.Added {
				revisionSource := schemaMapItemSource(operationsSources, info.operationItem.Revision, depSchemasDiff.Revision, name)
				result = append(result, info.newChange(RequestBodyDependentSchemaAddedId, []any{name}, "").
					WithSources(nil, revisionSource))
			}
			for _, name := range depSchemasDiff.Deleted {
				baseSource := schemaMapItemSource(operationsSources, info.operationItem.Base, depSchemasDiff.Base, name)
				result = append(result, info.newChange(RequestBodyDependentSchemaRemovedId, []any{name}, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.DependentSchemasDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)
			depSchemasDiff := p.propertyDiff.DependentSchemasDiff
			for _, name := range depSchemasDiff.Added {
				revisionSource := schemaMapItemSource(operationsSources, info.operationItem.Revision, depSchemasDiff.Revision, name)
				result = append(result, p.newChange(RequestPropertyDependentSchemaAddedId, []any{name, propName}, "").
					WithSources(nil, revisionSource))
			}
			for _, name := range depSchemasDiff.Deleted {
				baseSource := schemaMapItemSource(operationsSources, info.operationItem.Base, depSchemasDiff.Base, name)
				result = append(result, p.newChange(RequestPropertyDependentSchemaRemovedId, []any{name, propName}, "").
					WithSources(baseSource, nil))
			}
		})
	})

	return result
}
