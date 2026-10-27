package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertySunsetDeletedId             = "request-property-sunset-deleted"
	RequestPropertySunsetDateChangedTooSmallId = "request-property-sunset-date-changed-too-small"
)

func RequestPropertySunsetChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		stability, err := getStabilityLevel(info.operationItem.Revision.Extensions)
		if err != nil {
			// handled in CheckBackwardCompatibility
			return
		}

		deprecationDays := getDeprecationDays(config, stability)

		info.walkProperties(func(p propertyInfo) {
			// a property deprecated or reactivated in this change is reported by
			// RequestPropertyDeprecationCheck
			stillDeprecated := p.propertyDiff.Revision.Deprecated && p.propertyDiff.DeprecatedDiff == nil
			if !stillDeprecated {
				return
			}

			id, extra, ok := sunsetChangeId(p.propertyDiff.ExtensionsDiff, p.propertyDiff.Base.Extensions, p.propertyDiff.Revision.Extensions, deprecationDays, sunsetChangeIds{
				deleted:         RequestPropertySunsetDeletedId,
				sunsetParse:     RequestPropertyDeprecatedInvalidId,
				changedTooSmall: RequestPropertySunsetDateChangedTooSmallId,
			})
			if !ok {
				return
			}
			args := append([]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName)}, extra...)

			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, diff.SunsetExtension)
			result = append(result, p.newChange(id, args, "").WithSources(baseSource, revisionSource))
		})
	})

	return result
}
