package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyAdditionalPropertiesNarrowedId     = "request-body-additional-properties-narrowed"
	RequestBodyAdditionalPropertiesWidenedId      = "request-body-additional-properties-widened"
	RequestPropertyAdditionalPropertiesNarrowedId = "request-property-additional-properties-narrowed"
	RequestPropertyAdditionalPropertiesWidenedId  = "request-property-additional-properties-widened"
)

// additionalPropertiesMode orders the three forms of additionalProperties by
// how many extra keys they let through: any (absent or true), only keys
// matching a subschema, or none (false).
type additionalPropertiesMode int

const (
	additionalPropertiesAny additionalPropertiesMode = iota
	additionalPropertiesSchema
	additionalPropertiesNone
)

func getAdditionalPropertiesMode(schema *openapi3.Schema) additionalPropertiesMode {
	if schema == nil {
		return additionalPropertiesAny
	}
	if schema.AdditionalProperties.Schema != nil {
		return additionalPropertiesSchema
	}
	if schema.AdditionalProperties.Has != nil && !*schema.AdditionalProperties.Has {
		return additionalPropertiesNone
	}
	return additionalPropertiesAny
}

func (m additionalPropertiesMode) String() string {
	switch m {
	case additionalPropertiesSchema:
		return "schema"
	case additionalPropertiesNone:
		return "false"
	}
	return "true"
}

// additionalPropertiesChange reports whether the set of extra keys an object
// admits shrank (narrowed) or grew (widened). Edits inside an existing
// additionalProperties subschema are left to the subschema traversal.
func additionalPropertiesChange(schemaDiff *diff.SchemaDiff) (from, to additionalPropertiesMode, narrowed, widened bool) {
	if schemaDiff == nil || (schemaDiff.AdditionalPropertiesAllowedDiff == nil &&
		(schemaDiff.AdditionalPropertiesDiff == nil || (!schemaDiff.AdditionalPropertiesDiff.SchemaAdded && !schemaDiff.AdditionalPropertiesDiff.SchemaDeleted))) {
		return
	}
	from = getAdditionalPropertiesMode(schemaDiff.Base)
	to = getAdditionalPropertiesMode(schemaDiff.Revision)
	return from, to, to > from, to < from
}

func RequestPropertyAdditionalPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if from, to, narrowed, widened := additionalPropertiesChange(info.schemaDiff); narrowed || widened {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "additionalProperties")
			id := RequestBodyAdditionalPropertiesWidenedId
			if narrowed {
				id = RequestBodyAdditionalPropertiesNarrowedId
			}
			result = append(result, info.newChange(id, []any{from.String(), to.String()}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			from, to, narrowed, widened := additionalPropertiesChange(p.propertyDiff)
			if !narrowed && !widened {
				return
			}
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "additionalProperties")
			id := RequestPropertyAdditionalPropertiesWidenedId
			if narrowed {
				id = RequestPropertyAdditionalPropertiesNarrowedId
			}
			result = append(result, p.newChange(id, []any{propName, from.String(), to.String()}, "").
				WithSources(baseSource, revisionSource))
		})
	})

	return result
}
