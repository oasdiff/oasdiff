package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
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
	RequestRequiredPropertyBecameNonReadOnly31Id     = "request-required-property-became-not-read-only-in-openapi-31"
	RequiredInRequestsCommentId                      = "required-in-requests-comment"
	RequestPropertyBecameReadOnlyCommentId           = "request-property-became-read-only-comment"
)

func RequestPropertyWriteOnlyReadOnlyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if p.parent.Revision.Properties[p.propertyName] == nil {
				// removed properties processed by the RequestOptionalPropertyUpdatedCheck check
				return
			}
			// a property added to or removed from required is reported by
			// RequestPropertyRequiredUpdatedCheck
			requiredOnBothSides := slices.Contains(p.parent.Base.Required, p.propertyName) &&
				slices.Contains(p.parent.Revision.Required, p.propertyName)
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)

			if writeOnlyDiff := p.propertyDiff.WriteOnlyDiff; writeOnlyDiff != nil {
				propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "writeOnly")
				var id string
				if requiredOnBothSides {
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
				propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "readOnly")
				var id, comment string
				if requiredOnBothSides {
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
				if readOnlyDiff.To == true {
					comment = RequestPropertyBecameReadOnlyCommentId
				} else if requiredOnBothSides && openAPI31OrLater(diffReport.BaseOpenAPI) {
					// 3.0 applies required to a readOnly property in responses
					// only, 3.1 and later in requests too, so in 3.1 the
					// property was already required in requests
					id = RequestRequiredPropertyBecameNonReadOnly31Id
					comment = RequiredInRequestsCommentId
				}
				// the change is the property becoming or ceasing to be
				// readOnly, so readOnly cannot excuse it
				result = append(result, p.newChange(
					id,
					[]any{propName},
					comment,
				).withoutGuard(GuardReadOnly).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}

func openAPI31OrLater(version string) bool {
	return (&openapi3.T{OpenAPI: version}).IsOpenAPI31OrLater()
}
