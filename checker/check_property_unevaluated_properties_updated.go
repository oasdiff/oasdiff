package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyUnevaluatedPropertiesNarrowedId      = "request-body-unevaluated-properties-narrowed"
	RequestBodyUnevaluatedPropertiesWidenedId       = "request-body-unevaluated-properties-widened"
	RequestPropertyUnevaluatedPropertiesNarrowedId  = "request-property-unevaluated-properties-narrowed"
	RequestPropertyUnevaluatedPropertiesWidenedId   = "request-property-unevaluated-properties-widened"
	ResponseBodyUnevaluatedPropertiesNarrowedId     = "response-body-unevaluated-properties-narrowed"
	ResponseBodyUnevaluatedPropertiesWidenedId      = "response-body-unevaluated-properties-widened"
	ResponsePropertyUnevaluatedPropertiesNarrowedId = "response-property-unevaluated-properties-narrowed"
	ResponsePropertyUnevaluatedPropertiesWidenedId  = "response-property-unevaluated-properties-widened"
)

// getUnevaluatedPropertiesMode orders unevaluatedProperties the same way as
// additionalProperties: any (absent or true), a subschema, or none (false).
func getUnevaluatedPropertiesMode(schema *openapi3.Schema) additionalPropertiesMode {
	if schema == nil {
		return additionalPropertiesAny
	}
	if schema.UnevaluatedProperties.Schema != nil {
		return additionalPropertiesSchema
	}
	if schema.UnevaluatedProperties.Has != nil && !*schema.UnevaluatedProperties.Has {
		return additionalPropertiesNone
	}
	return additionalPropertiesAny
}

// unevaluatedPropertiesChange reports whether the set of unevaluated keys an
// object admits shrank (narrowed) or grew (widened) through
// unevaluatedProperties: false appearing or disappearing. A subschema appearing
// or disappearing against an open object is left to the
// unevaluated-properties-added/removed rules, and edits inside an existing
// subschema to the subschema traversal.
func unevaluatedPropertiesChange(schemaDiff *diff.SchemaDiff) (from, to additionalPropertiesMode, narrowed, widened bool) {
	if schemaDiff == nil || (schemaDiff.UnevaluatedPropertiesAllowedDiff == nil &&
		(schemaDiff.UnevaluatedPropertiesDiff == nil || (!schemaDiff.UnevaluatedPropertiesDiff.SchemaAdded && !schemaDiff.UnevaluatedPropertiesDiff.SchemaDeleted))) {
		return
	}
	from = getUnevaluatedPropertiesMode(schemaDiff.Base)
	to = getUnevaluatedPropertiesMode(schemaDiff.Revision)
	if from != additionalPropertiesNone && to != additionalPropertiesNone {
		return from, to, false, false
	}
	return from, to, to > from, to < from
}

// unevaluatedPropertiesClosedOnEitherSide reports whether the base or revision
// has unevaluatedProperties: false, in which case a subschema appearing or
// disappearing is reported as narrowed/widened rather than added/removed.
func unevaluatedPropertiesClosedOnEitherSide(schemaDiff *diff.SchemaDiff) bool {
	return schemaDiff != nil &&
		(getUnevaluatedPropertiesMode(schemaDiff.Base) == additionalPropertiesNone ||
			getUnevaluatedPropertiesMode(schemaDiff.Revision) == additionalPropertiesNone)
}

func RequestPropertyUnevaluatedPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if from, to, narrowed, widened := unevaluatedPropertiesChange(info.schemaDiff); narrowed || widened {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "unevaluatedProperties")
			id := RequestBodyUnevaluatedPropertiesWidenedId
			if narrowed {
				id = RequestBodyUnevaluatedPropertiesNarrowedId
			}
			result = append(result, info.newChange(id, []any{from.String(), to.String()}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			from, to, narrowed, widened := unevaluatedPropertiesChange(p.propertyDiff)
			if !narrowed && !widened {
				return
			}
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "unevaluatedProperties")
			id := RequestPropertyUnevaluatedPropertiesWidenedId
			if narrowed {
				id = RequestPropertyUnevaluatedPropertiesNarrowedId
			}
			result = append(result, p.newChange(id, []any{propName, from.String(), to.String()}, "").
				WithSources(baseSource, revisionSource))
		})
	})

	return result
}

func ResponsePropertyUnevaluatedPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if from, to, narrowed, widened := unevaluatedPropertiesChange(info.schemaDiff); narrowed || widened {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "unevaluatedProperties")
			id := ResponseBodyUnevaluatedPropertiesWidenedId
			if narrowed {
				id = ResponseBodyUnevaluatedPropertiesNarrowedId
			}
			result = append(result, info.newChange(id, []any{from.String(), to.String(), info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			from, to, narrowed, widened := unevaluatedPropertiesChange(p.propertyDiff)
			if !narrowed && !widened {
				return
			}
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "unevaluatedProperties")
			id := ResponsePropertyUnevaluatedPropertiesWidenedId
			if narrowed {
				id = ResponsePropertyUnevaluatedPropertiesNarrowedId
			}
			result = append(result, p.newChange(id, []any{propName, from.String(), to.String(), info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		})
	})

	return result
}
