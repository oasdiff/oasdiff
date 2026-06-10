package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyTypeGeneralizedId     = "request-body-type-generalized"
	RequestBodyTypeChangedId         = "request-body-type-changed"
	RequestPropertyTypeGeneralizedId = "request-property-type-generalized"
	RequestPropertyTypeChangedId     = "request-property-type-changed"
)

func RequestPropertyTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		schemaDiff := info.schemaDiff
		typeDiff := schemaDiff.TypeDiff
		formatDiff := schemaDiff.FormatDiff

		if !typeDiff.Empty() || !formatDiff.Empty() {
			// Body-level suppression also skips the property walk for this
			// media type (the pre-migration code used a `continue` in the
			// media-type for-loop, which had that effect). Preserved.
			if shouldSuppressTypeChangedForListOfTypes(schemaDiff) {
				return
			}
			// Suppress null-only type changes (handled by nullable checkers).
			if isNullTypeChange(typeDiff) && formatDiff.Empty() {
				return
			}

			id := RequestBodyTypeGeneralizedId
			if !isRequestTypeGeneralization(typeDiff, schemaDiff) && breakingTypeFormatChangedInRequestProperty(typeDiff, formatDiff, info.mediaType, schemaDiff) {
				id = RequestBodyTypeChangedId
			}
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, schemaDiff, "type")
			result = append(result, info.newChange(
				id,
				[]any{getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff)},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.Revision == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
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

			if !propTypeDiff.Empty() || !propFormatDiff.Empty() {
				id := RequestPropertyTypeGeneralizedId
				if !isRequestTypeGeneralization(propTypeDiff, propSchemaDiff) && breakingTypeFormatChangedInRequestProperty(propTypeDiff, propFormatDiff, info.mediaType, propSchemaDiff) {
					id = RequestPropertyTypeChangedId
				}

				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName), getBaseType(propSchemaDiff), getBaseFormat(propSchemaDiff), getRevisionType(propSchemaDiff), getRevisionFormat(propSchemaDiff)},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
