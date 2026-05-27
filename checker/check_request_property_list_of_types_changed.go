package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyListOfTypesWidenedId      = "request-body-list-of-types-widened"
	RequestBodyListOfTypesNarrowedId     = "request-body-list-of-types-narrowed"
	RequestPropertyListOfTypesWidenedId  = "request-property-list-of-types-widened"
	RequestPropertyListOfTypesNarrowedId = "request-property-list-of-types-narrowed"
)

func RequestPropertyListOfTypesChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		opInfo := newOpInfoFromDiff(config, info.operationItem, operationsSources, info.method, info.path)

		// Body-level
		result = append(result, checkBodyListOfTypesChange(
			opInfo,
			info.schemaDiff,
			info.mediaType,
			"",   // responseStatus not applicable for requests
			true, // isRequest
		)...)

		// Property-level
		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.Revision == nil || p.propertyDiff.Revision.ReadOnly {
				return
			}
			result = append(result, checkPropertyListOfTypesChange(
				opInfo,
				p.propertyPath,
				p.propertyName,
				p.propertyDiff,
				info.mediaType,
				"",   // responseStatus not applicable for requests
				true, // isRequest
			)...)
		})
	})

	return result
}
