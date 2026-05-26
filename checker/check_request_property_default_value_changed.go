package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyDefaultValueAddedId       = "request-body-default-value-added"
	RequestBodyDefaultValueRemovedId     = "request-body-default-value-removed"
	RequestBodyDefaultValueChangedId     = "request-body-default-value-changed"
	RequestPropertyDefaultValueAddedId   = "request-property-default-value-added"
	RequestPropertyDefaultValueRemovedId = "request-property-default-value-removed"
	RequestPropertyDefaultValueChangedId = "request-property-default-value-changed"
)

func RequestPropertyDefaultValueChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.DefaultDiff != nil {
			defaultValueDiff := info.schemaDiff.DefaultDiff
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "default")
			append1 := func(messageId string, a ...any) {
				result = append(result, info.newChange(messageId, a, "").WithSources(baseSource, revisionSource))
			}
			if defaultValueDiff.From == nil {
				append1(RequestBodyDefaultValueAddedId, info.mediaType, defaultValueDiff.To)
			} else if defaultValueDiff.To == nil {
				append1(RequestBodyDefaultValueRemovedId, info.mediaType, defaultValueDiff.From)
			} else {
				append1(RequestBodyDefaultValueChangedId, info.mediaType, defaultValueDiff.From, defaultValueDiff.To)
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff == nil || p.propertyDiff.DefaultDiff == nil {
				return
			}

			defaultValueDiff := p.propertyDiff.DefaultDiff
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "default")
			appendProp := func(messageId string, a ...any) {
				result = append(result, p.newChange(messageId, a, "").WithSources(propBaseSource, propRevisionSource))
			}

			if defaultValueDiff.From == nil {
				appendProp(RequestPropertyDefaultValueAddedId, p.propertyName, defaultValueDiff.To)
			} else if defaultValueDiff.To == nil {
				appendProp(RequestPropertyDefaultValueRemovedId, p.propertyName, defaultValueDiff.From)
			} else {
				appendProp(RequestPropertyDefaultValueChangedId, p.propertyName, defaultValueDiff.From, defaultValueDiff.To)
			}
		})
	})

	return result
}
