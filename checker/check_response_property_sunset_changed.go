package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertySunsetDeletedId             = "response-property-sunset-deleted"
	ResponsePropertySunsetDateChangedTooSmallId = "response-property-sunset-date-changed-too-small"
)

func ResponsePropertySunsetChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		stability, err := getStabilityLevel(info.operationItem.Revision.Extensions)
		if err != nil {
			// handled in CheckBackwardCompatibility
			return
		}

		deprecationDays := getDeprecationDays(config, stability)

		info.walkProperties(func(p propertyInfo) {
			// a property deprecated or reactivated in this change is reported by
			// ResponsePropertyDeprecationCheck
			stillDeprecated := p.propertyDiff.Revision.Deprecated && p.propertyDiff.DeprecatedDiff == nil
			if !stillDeprecated {
				return
			}

			id, extra, ok := sunsetChangeId(p.propertyDiff.ExtensionsDiff, p.propertyDiff.Base.Extensions, p.propertyDiff.Revision.Extensions, deprecationDays, sunsetChangeIds{
				deleted:         ResponsePropertySunsetDeletedId,
				sunsetParse:     ResponsePropertyDeprecatedInvalidId,
				changedTooSmall: ResponsePropertySunsetDateChangedTooSmallId,
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
