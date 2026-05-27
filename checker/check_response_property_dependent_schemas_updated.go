package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyDependentSchemaAddedId       = "response-body-dependent-schema-added"
	ResponseBodyDependentSchemaRemovedId     = "response-body-dependent-schema-removed"
	ResponsePropertyDependentSchemaAddedId   = "response-property-dependent-schema-added"
	ResponsePropertyDependentSchemaRemovedId = "response-property-dependent-schema-removed"
)

func ResponsePropertyDependentSchemasUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.DependentSchemasDiff != nil {
			depSchemasDiff := info.schemaDiff.DependentSchemasDiff
			for _, name := range depSchemasDiff.Added {
				revisionSource := SchemaMapItemSource(operationsSources, info.operationItem.Revision, depSchemasDiff.Revision, name)
				result = append(result, info.newChange(ResponseBodyDependentSchemaAddedId, []any{name, info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			for _, name := range depSchemasDiff.Deleted {
				baseSource := SchemaMapItemSource(operationsSources, info.operationItem.Base, depSchemasDiff.Base, name)
				result = append(result, info.newChange(ResponseBodyDependentSchemaRemovedId, []any{name, info.responseStatus}, "").
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
				revisionSource := SchemaMapItemSource(operationsSources, info.operationItem.Revision, depSchemasDiff.Revision, name)
				result = append(result, p.newChange(ResponsePropertyDependentSchemaAddedId, []any{name, propName, info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			for _, name := range depSchemasDiff.Deleted {
				baseSource := SchemaMapItemSource(operationsSources, info.operationItem.Base, depSchemasDiff.Base, name)
				result = append(result, p.newChange(ResponsePropertyDependentSchemaRemovedId, []any{name, propName, info.responseStatus}, "").
					WithSources(baseSource, nil))
			}
		})
	})

	return result
}
