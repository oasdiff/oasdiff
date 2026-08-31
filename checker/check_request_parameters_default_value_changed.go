package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterDefaultValueChangedId = "request-parameter-default-value-changed"
	RequestParameterDefaultValueAddedId   = "request-parameter-default-value-added"
	RequestParameterDefaultValueRemovedId = "request-parameter-default-value-removed"
)

func RequestParameterDefaultValueChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {

		baseParam := p.paramDiff.Base
		if baseParam == nil || baseParam.Required {
			return
		}

		revisionParam := p.paramDiff.Revision
		if revisionParam == nil || revisionParam.Required {
			return
		}

		if p.paramDiff.SchemaDiff == nil {
			return
		}

		defaultValueDiff := p.paramDiff.SchemaDiff.DefaultDiff
		if defaultValueDiff.Empty() {
			return
		}

		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "default")
		appendResultItem := func(messageId string, a ...any) {
			result = append(result, p.opInfo.NewApiChange(
				messageId,
				a,
				"",
			).WithSources(baseSource, revisionSource))
		}

		if defaultValueDiff.From == nil {
			appendResultItem(RequestParameterDefaultValueAddedId, p.location, p.name, defaultValueDiff.To)
		} else if defaultValueDiff.To == nil {
			appendResultItem(RequestParameterDefaultValueRemovedId, p.location, p.name, defaultValueDiff.From)
		} else {
			appendResultItem(RequestParameterDefaultValueChangedId, p.location, p.name, defaultValueDiff.From, defaultValueDiff.To)
		}
	})
	return result
}
