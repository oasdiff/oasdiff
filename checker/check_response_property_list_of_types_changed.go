package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyListOfTypesWidenedId      = "response-body-list-of-types-widened"
	ResponseBodyListOfTypesNarrowedId     = "response-body-list-of-types-narrowed"
	ResponsePropertyListOfTypesWidenedId  = "response-property-list-of-types-widened"
	ResponsePropertyListOfTypesNarrowedId = "response-property-list-of-types-narrowed"
)

func ResponsePropertyListOfTypesChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		opInfo := newOpInfoFromDiff(config, info.operationItem, operationsSources, info.method, info.path)

		// Body-level
		result = append(result, checkBodyListOfTypesChange(
			opInfo,
			info.schemaDiff,
			info.mediaType,
			info.responseStatus,
			false, // isRequest
		)...)

		// Property-level
		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff == nil || p.propertyDiff.Revision == nil {
				return
			}
			result = append(result, checkPropertyListOfTypesChange(
				opInfo,
				p.propertyPath,
				p.propertyName,
				p.propertyDiff,
				info.mediaType,
				info.responseStatus,
				false, // isRequest
			)...)
		})
	})

	return result
}
