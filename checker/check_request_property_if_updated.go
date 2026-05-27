package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyIfAddedId       = "request-body-if-added"
	RequestBodyIfRemovedId     = "request-body-if-removed"
	RequestBodyThenAddedId     = "request-body-then-added"
	RequestBodyThenRemovedId   = "request-body-then-removed"
	RequestBodyElseAddedId     = "request-body-else-added"
	RequestBodyElseRemovedId   = "request-body-else-removed"
	RequestPropertyIfAddedId   = "request-property-if-added"
	RequestPropertyIfRemovedId = "request-property-if-removed"

	RequestPropertyThenAddedId   = "request-property-then-added"
	RequestPropertyThenRemovedId = "request-property-then-removed"
	RequestPropertyElseAddedId   = "request-property-else-added"
	RequestPropertyElseRemovedId = "request-property-else-removed"
)

func RequestPropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		for _, entry := range []struct {
			schemaDiff *diff.SchemaDiff
			addedId    string
			removedId  string
			field      string
		}{
			{info.schemaDiff.IfDiff, RequestBodyIfAddedId, RequestBodyIfRemovedId, "if"},
			{info.schemaDiff.ThenDiff, RequestBodyThenAddedId, RequestBodyThenRemovedId, "then"},
			{info.schemaDiff.ElseDiff, RequestBodyElseAddedId, RequestBodyElseRemovedId, "else"},
		} {
			if entry.schemaDiff == nil {
				continue
			}
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, entry.field)
			if entry.schemaDiff.SchemaAdded {
				result = append(result, info.newChange(entry.addedId, nil, "").
					WithSources(nil, revisionSource))
			}
			if entry.schemaDiff.SchemaDeleted {
				result = append(result, info.newChange(entry.removedId, nil, "").
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
				{p.propertyDiff.IfDiff, RequestPropertyIfAddedId, RequestPropertyIfRemovedId, "if"},
				{p.propertyDiff.ThenDiff, RequestPropertyThenAddedId, RequestPropertyThenRemovedId, "then"},
				{p.propertyDiff.ElseDiff, RequestPropertyElseAddedId, RequestPropertyElseRemovedId, "else"},
			} {
				if entry.schemaDiff == nil {
					continue
				}
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, entry.field)
				if entry.schemaDiff.SchemaAdded {
					result = append(result, p.newChange(entry.addedId, []any{propName}, "").
						WithSources(nil, propRevisionSource))
				}
				if entry.schemaDiff.SchemaDeleted {
					result = append(result, p.newChange(entry.removedId, []any{propName}, "").
						WithSources(propBaseSource, nil))
				}
			}
		})
	})

	return result
}
