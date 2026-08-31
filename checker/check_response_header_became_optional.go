package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseHeaderBecameOptionalId = "response-header-became-optional"
)

func ResponseHeaderBecameOptionalCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedResponseHeaders(diffReport, operationsSources, config, func(h headerInfo) {
		requiredDiff := h.headerDiff.RequiredDiff
		if requiredDiff == nil {
			return
		}
		if requiredDiff.From != true {
			return
		}

		baseSource, revisionSource := headerSources(operationsSources, h.opInfo.methodDiff, h.responseDiff, h.name)
		result = append(result, h.opInfo.NewApiChange(
			ResponseHeaderBecameOptionalId,
			[]any{h.name, h.responseStatus},
			"",
		).WithSources(baseSource, revisionSource))
	})
	return result
}
