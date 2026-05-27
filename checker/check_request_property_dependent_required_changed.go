package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyDependentRequiredAddedId       = "request-body-dependent-required-added"
	RequestBodyDependentRequiredRemovedId     = "request-body-dependent-required-removed"
	RequestBodyDependentRequiredChangedId     = "request-body-dependent-required-changed"
	RequestPropertyDependentRequiredAddedId   = "request-property-dependent-required-added"
	RequestPropertyDependentRequiredRemovedId = "request-property-dependent-required-removed"
	RequestPropertyDependentRequiredChangedId = "request-property-dependent-required-changed"
)

func RequestPropertyDependentRequiredChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.DependentRequiredDiff != nil {
			depReqDiff := info.schemaDiff.DependentRequiredDiff
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "dependentRequired")
			for key, values := range depReqDiff.Added {
				result = append(result, info.newChange(
					RequestBodyDependentRequiredAddedId,
					[]any{key, strings.Join(values, ", ")},
					"",
				).WithSources(nil, revisionSource))
			}
			for key, values := range depReqDiff.Deleted {
				result = append(result, info.newChange(
					RequestBodyDependentRequiredRemovedId,
					[]any{key, strings.Join(values, ", ")},
					"",
				).WithSources(baseSource, nil))
			}
			for key, stringsDiff := range depReqDiff.Modified {
				result = append(result, info.newChange(
					RequestBodyDependentRequiredChangedId,
					[]any{key, formatDependentRequiredModification(stringsDiff)},
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
					RequestPropertyDependentRequiredAddedId,
					[]any{propName, key, strings.Join(values, ", ")},
					"",
				).WithSources(nil, propRevisionSource))
			}
			for key, values := range depReqDiff.Deleted {
				result = append(result, p.newChange(
					RequestPropertyDependentRequiredRemovedId,
					[]any{propName, key, strings.Join(values, ", ")},
					"",
				).WithSources(propBaseSource, nil))
			}
			for key, stringsDiff := range depReqDiff.Modified {
				result = append(result, p.newChange(
					RequestPropertyDependentRequiredChangedId,
					[]any{propName, key, formatDependentRequiredModification(stringsDiff)},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
