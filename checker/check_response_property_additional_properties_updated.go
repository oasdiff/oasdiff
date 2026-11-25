package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyAdditionalPropertiesNarrowedId     = "response-body-additional-properties-narrowed"
	ResponseBodyAdditionalPropertiesWidenedId      = "response-body-additional-properties-widened"
	ResponsePropertyAdditionalPropertiesNarrowedId = "response-property-additional-properties-narrowed"
	ResponsePropertyAdditionalPropertiesWidenedId  = "response-property-additional-properties-widened"
)

func ResponsePropertyAdditionalPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if from, to, narrowed, widened := additionalPropertiesChange(info.schemaDiff); narrowed || widened {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "additionalProperties")
			id := ResponseBodyAdditionalPropertiesWidenedId
			if narrowed {
				id = ResponseBodyAdditionalPropertiesNarrowedId
			}
			result = append(result, info.newChange(id, []any{from.String(), to.String(), info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			from, to, narrowed, widened := additionalPropertiesChange(p.propertyDiff)
			if !narrowed && !widened {
				return
			}
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "additionalProperties")
			id := ResponsePropertyAdditionalPropertiesWidenedId
			if narrowed {
				id = ResponsePropertyAdditionalPropertiesNarrowedId
			}
			result = append(result, p.newChange(id, []any{propName, from.String(), to.String(), info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		})
	})

	return result
}
