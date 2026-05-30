package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseOptionalPropertyBecameNonWriteOnlyId = "response-optional-property-became-not-write-only"
	ResponseOptionalPropertyBecameWriteOnlyId    = "response-optional-property-became-write-only"
	ResponseOptionalPropertyBecameReadOnlyId     = "response-optional-property-became-read-only"
	ResponseOptionalPropertyBecameNonReadOnlyId  = "response-optional-property-became-not-read-only"
)

func ResponseOptionalPropertyWriteOnlyReadOnlyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if p.parent.Revision.Properties[p.propertyName] == nil {
				// removed properties processed by the ResponseOptionalPropertyUpdatedCheck check
				return
			}
			if slices.Contains(p.parent.Base.Required, p.propertyName) {
				// skip required properties — checked at ResponseRequiredPropertyWriteOnlyReadOnlyCheck
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)

			if writeOnlyDiff := p.propertyDiff.WriteOnlyDiff; writeOnlyDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "writeOnly")
				id := ResponseOptionalPropertyBecameNonWriteOnlyId
				if writeOnlyDiff.To == true {
					id = ResponseOptionalPropertyBecameWriteOnlyId
				}
				result = append(result, p.newChange(
					id,
					[]any{propName, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}

			if readOnlyDiff := p.propertyDiff.ReadOnlyDiff; readOnlyDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "readOnly")
				id := ResponseOptionalPropertyBecameNonReadOnlyId
				if readOnlyDiff.To == true {
					id = ResponseOptionalPropertyBecameReadOnlyId
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
