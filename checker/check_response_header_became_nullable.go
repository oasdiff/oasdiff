package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseHeaderBecameNullableId    = "response-header-became-nullable"
	ResponseHeaderBecameNotNullableId = "response-header-became-not-nullable"
)

// ResponseHeaderBecameNullableCheck reports a nullability change on a response
// header's schema, the response-header mirror of ResponsePropertyBecameNullableCheck.
// A response header becoming nullable means a client that never expected a null
// value can now receive one (breaking); becoming not-nullable narrows the
// server's output and is safe (info).
func ResponseHeaderBecameNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedResponseHeaders(diffReport, operationsSources, config, func(h headerInfo) {
		if id := nullabilityChangeId(h.headerDiff.SchemaDiff, ResponseHeaderBecameNullableId, ResponseHeaderBecameNotNullableId); id != "" {
			baseSource, revisionSource := headerSources(operationsSources, h.opInfo.methodDiff, h.responseDiff, h.name)
			result = append(result, h.opInfo.NewApiChange(
				id,
				[]any{h.name, h.responseStatus},
				"",
			).WithSources(baseSource, revisionSource))
		}
	})
	return result
}
