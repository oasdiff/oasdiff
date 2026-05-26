package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyConstAddedId       = "request-body-const-added"
	RequestBodyConstRemovedId     = "request-body-const-removed"
	RequestBodyConstChangedId     = "request-body-const-changed"
	RequestPropertyConstAddedId   = "request-property-const-added"
	RequestPropertyConstRemovedId = "request-property-const-removed"
	RequestPropertyConstChangedId = "request-property-const-changed"
)

func RequestPropertyConstChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.ConstDiff != nil {
			constDiff := info.schemaDiff.ConstDiff
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "const")

			if constDiff.From == nil {
				result = append(result, info.newChange(
					RequestBodyConstAddedId,
					[]any{info.mediaType, constDiff.To},
					"",
				).WithSources(nil, revisionSource))
			} else if constDiff.To == nil {
				result = append(result, info.newChange(
					RequestBodyConstRemovedId,
					[]any{info.mediaType, constDiff.From},
					"",
				).WithSources(baseSource, nil))
			} else {
				result = append(result, info.newChange(
					RequestBodyConstChangedId,
					[]any{info.mediaType, constDiff.From, constDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff == nil || p.propertyDiff.ConstDiff == nil {
				return
			}

			constDiff := p.propertyDiff.ConstDiff
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "const")

			if constDiff.From == nil {
				result = append(result, p.newChange(
					RequestPropertyConstAddedId,
					[]any{p.propertyName, constDiff.To},
					"",
				).WithSources(nil, propRevisionSource))
			} else if constDiff.To == nil {
				result = append(result, p.newChange(
					RequestPropertyConstRemovedId,
					[]any{p.propertyName, constDiff.From},
					"",
				).WithSources(propBaseSource, nil))
			} else {
				result = append(result, p.newChange(
					RequestPropertyConstChangedId,
					[]any{p.propertyName, constDiff.From, constDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
