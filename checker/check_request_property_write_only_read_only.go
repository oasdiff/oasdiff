package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestOptionalPropertyBecameNonWriteOnlyCheckId = "request-optional-property-became-not-write-only"
	RequestOptionalPropertyBecameWriteOnlyCheckId    = "request-optional-property-became-write-only"
	RequestOptionalPropertyBecameReadOnlyCheckId     = "request-optional-property-became-read-only"
	RequestOptionalPropertyBecameNonReadOnlyCheckId  = "request-optional-property-became-not-read-only"
	RequestRequiredPropertyBecameNonWriteOnlyCheckId = "request-required-property-became-not-write-only"
	RequestRequiredPropertyBecameWriteOnlyCheckId    = "request-required-property-became-write-only"
	RequestRequiredPropertyBecameReadOnlyCheckId     = "request-required-property-became-read-only"
	RequestRequiredPropertyBecameNonReadOnlyCheckId  = "request-required-property-became-not-read-only"
)

func RequestPropertyWriteOnlyReadOnlyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if p.parent.Revision.Properties[p.propertyName] == nil {
				// removed properties processed by the RequestOptionalPropertyUpdatedCheck check
				return
			}
			required := slices.Contains(p.parent.Base.Required, p.propertyName)
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if writeOnlyDiff := p.propertyDiff.WriteOnlyDiff; writeOnlyDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "writeOnly")
				var id string
				if required {
					id = RequestRequiredPropertyBecameNonWriteOnlyCheckId
					if writeOnlyDiff.To == true {
						id = RequestRequiredPropertyBecameWriteOnlyCheckId
					}
				} else {
					id = RequestOptionalPropertyBecameNonWriteOnlyCheckId
					if writeOnlyDiff.To == true {
						id = RequestOptionalPropertyBecameWriteOnlyCheckId
					}
				}
				result = append(result, p.newChange(
					id,
					[]any{propName},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}

			if readOnlyDiff := p.propertyDiff.ReadOnlyDiff; readOnlyDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "readOnly")
				var id string
				if required {
					id = RequestRequiredPropertyBecameNonReadOnlyCheckId
					if readOnlyDiff.To == true {
						id = RequestRequiredPropertyBecameReadOnlyCheckId
					}
				} else {
					id = RequestOptionalPropertyBecameNonReadOnlyCheckId
					if readOnlyDiff.To == true {
						id = RequestOptionalPropertyBecameReadOnlyCheckId
					}
				}
				result = append(result, p.newChange(
					id,
					[]any{propName},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
