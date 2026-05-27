package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyDependentRequiredAddedId       = "response-body-dependent-required-added"
	ResponseBodyDependentRequiredRemovedId     = "response-body-dependent-required-removed"
	ResponseBodyDependentRequiredChangedId     = "response-body-dependent-required-changed"
	ResponsePropertyDependentRequiredAddedId   = "response-property-dependent-required-added"
	ResponsePropertyDependentRequiredRemovedId = "response-property-dependent-required-removed"
	ResponsePropertyDependentRequiredChangedId = "response-property-dependent-required-changed"
)

func ResponsePropertyDependentRequiredChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.DependentRequiredDiff != nil {
			depReqDiff := info.schemaDiff.DependentRequiredDiff
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "dependentRequired")
			for key, values := range depReqDiff.Added {
				result = append(result, info.newChange(
					ResponseBodyDependentRequiredAddedId,
					[]any{info.responseStatus, key, strings.Join(values, ", ")},
					"",
				).WithSources(nil, revisionSource))
			}
			for key, values := range depReqDiff.Deleted {
				result = append(result, info.newChange(
					ResponseBodyDependentRequiredRemovedId,
					[]any{info.responseStatus, key, strings.Join(values, ", ")},
					"",
				).WithSources(baseSource, nil))
			}
			for key, stringsDiff := range depReqDiff.Modified {
				result = append(result, info.newChange(
					ResponseBodyDependentRequiredChangedId,
					[]any{key, info.responseStatus, formatDependentRequiredModification(stringsDiff)},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.DependentRequiredDiff == nil {
				return
			}
			depReqDiff := p.propertyDiff.DependentRequiredDiff
			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "dependentRequired")
			for key, values := range depReqDiff.Added {
				result = append(result, p.newChange(
					ResponsePropertyDependentRequiredAddedId,
					[]any{propName, info.responseStatus, key, strings.Join(values, ", ")},
					"",
				).WithSources(nil, propRevisionSource))
			}
			for key, values := range depReqDiff.Deleted {
				result = append(result, p.newChange(
					ResponsePropertyDependentRequiredRemovedId,
					[]any{propName, info.responseStatus, key, strings.Join(values, ", ")},
					"",
				).WithSources(propBaseSource, nil))
			}
			for key, stringsDiff := range depReqDiff.Modified {
				result = append(result, p.newChange(
					ResponsePropertyDependentRequiredChangedId,
					[]any{propName, key, info.responseStatus, formatDependentRequiredModification(stringsDiff)},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
