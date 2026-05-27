package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyIfAddedId       = "response-body-if-added"
	ResponseBodyIfRemovedId     = "response-body-if-removed"
	ResponseBodyThenAddedId     = "response-body-then-added"
	ResponseBodyThenRemovedId   = "response-body-then-removed"
	ResponseBodyElseAddedId     = "response-body-else-added"
	ResponseBodyElseRemovedId   = "response-body-else-removed"
	ResponsePropertyIfAddedId   = "response-property-if-added"
	ResponsePropertyIfRemovedId = "response-property-if-removed"

	ResponsePropertyThenAddedId   = "response-property-then-added"
	ResponsePropertyThenRemovedId = "response-property-then-removed"
	ResponsePropertyElseAddedId   = "response-property-else-added"
	ResponsePropertyElseRemovedId = "response-property-else-removed"
)

func ResponsePropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		for _, entry := range []struct {
			schemaDiff *diff.SchemaDiff
			addedId    string
			removedId  string
			field      string
		}{
			{info.schemaDiff.IfDiff, ResponseBodyIfAddedId, ResponseBodyIfRemovedId, "if"},
			{info.schemaDiff.ThenDiff, ResponseBodyThenAddedId, ResponseBodyThenRemovedId, "then"},
			{info.schemaDiff.ElseDiff, ResponseBodyElseAddedId, ResponseBodyElseRemovedId, "else"},
		} {
			if entry.schemaDiff == nil {
				continue
			}
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, entry.field)
			if entry.schemaDiff.SchemaAdded {
				result = append(result, info.newChange(entry.addedId, []any{info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			if entry.schemaDiff.SchemaDeleted {
				result = append(result, info.newChange(entry.removedId, []any{info.responseStatus}, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			for _, entry := range []struct {
				schemaDiff *diff.SchemaDiff
				addedId    string
				removedId  string
				field      string
			}{
				{p.propertyDiff.IfDiff, ResponsePropertyIfAddedId, ResponsePropertyIfRemovedId, "if"},
				{p.propertyDiff.ThenDiff, ResponsePropertyThenAddedId, ResponsePropertyThenRemovedId, "then"},
				{p.propertyDiff.ElseDiff, ResponsePropertyElseAddedId, ResponsePropertyElseRemovedId, "else"},
			} {
				if entry.schemaDiff == nil {
					continue
				}
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, entry.field)
				if entry.schemaDiff.SchemaAdded {
					result = append(result, p.newChange(entry.addedId, []any{propName, info.responseStatus}, "").
						WithSources(nil, propRevisionSource))
				}
				if entry.schemaDiff.SchemaDeleted {
					result = append(result, p.newChange(entry.removedId, []any{propName, info.responseStatus}, "").
						WithSources(propBaseSource, nil))
				}
			}
		})
	})

	return result
}
