package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseRequiredPropertyBecameNonWriteOnlyId = "response-required-property-became-not-write-only"
	ResponseRequiredPropertyBecameWriteOnlyId    = "response-required-property-became-write-only"
	ResponseRequiredPropertyBecameReadOnlyId     = "response-required-property-became-read-only"
	ResponseRequiredPropertyBecameNonReadOnlyId  = "response-required-property-became-not-read-only"
)

func ResponseRequiredPropertyWriteOnlyReadOnlyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if p.parent.Revision.Properties[p.propertyName] == nil {
				// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
				return
			}
			if !slices.Contains(p.parent.Base.Required, p.propertyName) {
				// skip non-required properties
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)

			if writeOnlyDiff := p.propertyDiff.WriteOnlyDiff; writeOnlyDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "writeOnly")
				id := ResponseRequiredPropertyBecameNonWriteOnlyId
				comment := commentId(ResponseRequiredPropertyBecameNonWriteOnlyId)
				if writeOnlyDiff.To == true {
					id = ResponseRequiredPropertyBecameWriteOnlyId
					comment = ""
				}
				result = append(result, p.newChange(
					id,
					[]any{propName, info.responseStatus},
					comment,
				).WithSources(propBaseSource, propRevisionSource))
			}

			if readOnlyDiff := p.propertyDiff.ReadOnlyDiff; readOnlyDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "readOnly")
				id := ResponseRequiredPropertyBecameNonReadOnlyId
				if readOnlyDiff.To == true {
					id = ResponseRequiredPropertyBecameReadOnlyId
				}
				result = append(result, p.newChange(
					id,
					[]any{propName, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
