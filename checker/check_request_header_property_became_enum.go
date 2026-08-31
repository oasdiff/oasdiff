package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestHeaderPropertyBecameEnumId = "request-header-property-became-enum"
)

func RequestHeaderPropertyBecameEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.location != "header" {
			return
		}
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "enum")
		if p.paramDiff.SchemaDiff.EnumDiff != nil && p.paramDiff.SchemaDiff.EnumDiff.EnumAdded {
			result = append(result, p.opInfo.NewApiChange(
				RequestHeaderPropertyBecameEnumId,
				[]any{p.name},
				"",
			).WithSources(baseSource, revisionSource))
		}

		checkModifiedPropertiesDiff(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {

				if enumDiff := propertyDiff.EnumDiff; enumDiff == nil || !enumDiff.EnumAdded {
					return
				}

				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, propertyDiff, "enum")
				result = append(result, p.opInfo.NewApiChange(
					RequestHeaderPropertyBecameEnumId,
					[]any{p.name, propertyFullName(propertyPath, propertyName)},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			})
	})
	return result
}
