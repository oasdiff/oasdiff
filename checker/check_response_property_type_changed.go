package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyTypeChangedId     = "response-body-type-changed"
	ResponsePropertyTypeChangedId = "response-property-type-changed"
)

func ResponsePropertyTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		schemaDiff := info.schemaDiff

		// Body-level suppression also skips the property walk for this
		// media type.
		if shouldSuppressTypeChangedForListOfTypes(schemaDiff) {
			return
		}

		typeDiff := schemaDiff.TypeDiff
		formatDiff := schemaDiff.FormatDiff

		// Suppress null-only type changes (handled by nullable checkers).
		if isNullTypeChange(typeDiff) && formatDiff.Empty() {
			return
		}

		if responseTypeFormatBreaking(typeDiff, formatDiff, info.mediaType, schemaDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, schemaDiff, "type")
			result = append(result, info.newChange(
				ResponseBodyTypeChangedId,
				[]any{getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff), info.responseStatus},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff == nil || p.propertyDiff.Revision == nil {
				return
			}

			if shouldSuppressPropertyTypeChangedForListOfTypes(p.propertyDiff) {
				return
			}
			if isNullTypeChange(p.propertyDiff.TypeDiff) && p.propertyDiff.FormatDiff.Empty() {
				return
			}

			propSchemaDiff := p.propertyDiff
			propTypeDiff := propSchemaDiff.TypeDiff
			propFormatDiff := propSchemaDiff.FormatDiff

			if responseTypeFormatBreaking(propTypeDiff, propFormatDiff, info.mediaType, propSchemaDiff) {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					ResponsePropertyTypeChangedId,
					[]any{propertyFullName(p.propertyPath, p.propertyName), getBaseType(propSchemaDiff), getBaseFormat(propSchemaDiff), getRevisionType(propSchemaDiff), getRevisionFormat(propSchemaDiff), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
