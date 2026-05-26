package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyDefaultValueAddedId       = "response-body-default-value-added"
	ResponseBodyDefaultValueRemovedId     = "response-body-default-value-removed"
	ResponseBodyDefaultValueChangedId     = "response-body-default-value-changed"
	ResponsePropertyDefaultValueAddedId   = "response-property-default-value-added"
	ResponsePropertyDefaultValueRemovedId = "response-property-default-value-removed"
	ResponsePropertyDefaultValueChangedId = "response-property-default-value-changed"
)

func ResponsePropertyDefaultValueChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.DefaultDiff != nil {
			defaultValueDiff := info.schemaDiff.DefaultDiff
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "default")
			append1 := func(messageId string, a ...any) {
				result = append(result, info.newChange(messageId, a, "").WithSources(baseSource, revisionSource))
			}
			if defaultValueDiff.From == nil {
				append1(ResponseBodyDefaultValueAddedId, info.mediaType, defaultValueDiff.To, info.responseStatus)
			} else if defaultValueDiff.To == nil {
				append1(ResponseBodyDefaultValueRemovedId, info.mediaType, defaultValueDiff.From, info.responseStatus)
			} else {
				append1(ResponseBodyDefaultValueChangedId, info.mediaType, defaultValueDiff.From, defaultValueDiff.To, info.responseStatus)
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff == nil || p.propertyDiff.Revision == nil || p.propertyDiff.DefaultDiff == nil {
				return
			}

			defaultValueDiff := p.propertyDiff.DefaultDiff
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "default")
			appendProp := func(messageId string, a ...any) {
				result = append(result, p.newChange(messageId, a, "").WithSources(propBaseSource, propRevisionSource))
			}
			if defaultValueDiff.From == nil {
				appendProp(ResponsePropertyDefaultValueAddedId, p.propertyName, defaultValueDiff.To, info.responseStatus)
			} else if defaultValueDiff.To == nil {
				appendProp(ResponsePropertyDefaultValueRemovedId, p.propertyName, defaultValueDiff.From, info.responseStatus)
			} else {
				appendProp(ResponsePropertyDefaultValueChangedId, p.propertyName, defaultValueDiff.From, defaultValueDiff.To, info.responseStatus)
			}
		})
	})

	return result
}
