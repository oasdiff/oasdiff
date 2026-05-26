package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyConstAddedId       = "response-body-const-added"
	ResponseBodyConstRemovedId     = "response-body-const-removed"
	ResponseBodyConstChangedId     = "response-body-const-changed"
	ResponsePropertyConstAddedId   = "response-property-const-added"
	ResponsePropertyConstRemovedId = "response-property-const-removed"
	ResponsePropertyConstChangedId = "response-property-const-changed"
)

func ResponsePropertyConstChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.ConstDiff != nil {
			constDiff := info.schemaDiff.ConstDiff
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "const")

			if constDiff.From == nil {
				result = append(result, info.newChange(
					ResponseBodyConstAddedId,
					[]any{info.mediaType, constDiff.To, info.responseStatus},
					"",
				).WithSources(nil, revisionSource))
			} else if constDiff.To == nil {
				result = append(result, info.newChange(
					ResponseBodyConstRemovedId,
					[]any{info.mediaType, constDiff.From, info.responseStatus},
					"",
				).WithSources(baseSource, nil))
			} else {
				result = append(result, info.newChange(
					ResponseBodyConstChangedId,
					[]any{info.mediaType, constDiff.From, constDiff.To, info.responseStatus},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff == nil || p.propertyDiff.Revision == nil || p.propertyDiff.ConstDiff == nil {
				return
			}

			constDiff := p.propertyDiff.ConstDiff
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "const")

			if constDiff.From == nil {
				result = append(result, p.newChange(
					ResponsePropertyConstAddedId,
					[]any{p.propertyName, constDiff.To, info.responseStatus},
					"",
				).WithSources(nil, propRevisionSource))
			} else if constDiff.To == nil {
				result = append(result, p.newChange(
					ResponsePropertyConstRemovedId,
					[]any{p.propertyName, constDiff.From, info.responseStatus},
					"",
				).WithSources(propBaseSource, nil))
			} else {
				result = append(result, p.newChange(
					ResponsePropertyConstChangedId,
					[]any{p.propertyName, constDiff.From, constDiff.To, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
