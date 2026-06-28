package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyTypeChangedId         = "response-body-type-changed"
	ResponseBodyTypeGeneralizedId     = "response-body-type-generalized"
	ResponseBodyTypeSpecializedId     = "response-body-type-specialized"
	ResponsePropertyTypeChangedId     = "response-property-type-changed"
	ResponsePropertyTypeGeneralizedId = "response-property-type-generalized"
	ResponsePropertyTypeSpecializedId = "response-property-type-specialized"
)

// responseTypeChangeId classifies a response type/format change into one of three
// verdicts, the contravariant mirror of the request side:
//   - specialized (info): the returned type set narrowed (e.g. number -> integer),
//     so clients keep getting values they already handled. Safe.
//   - generalized (error): the returned type set widened (e.g. integer -> number),
//     so the server may now return a type the client did not expect. Breaking.
//   - changed (error): a genuinely incompatible change (e.g. string -> integer).
func responseTypeChangeId(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff, specializedId, generalizedId, changedId string) string {
	if !responseTypeFormatBreaking(typeDiff, formatDiff, mediaType, schemaDiff) {
		return specializedId
	}
	if isResponseTypeWidening(typeDiff, schemaDiff) {
		return generalizedId
	}
	return changedId
}

func ResponsePropertyTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		schemaDiff := info.schemaDiff
		typeDiff := schemaDiff.TypeDiff
		formatDiff := schemaDiff.FormatDiff

		if !typeDiff.Empty() || !formatDiff.Empty() {
			// Body-level suppression also skips the property walk for this
			// media type.
			if shouldSuppressTypeChangedForListOfTypes(schemaDiff) {
				return
			}
			// A oneOf wrapping (#702) reads as a top-level type change to "any"
			// (the oneOf wrapper has no type of its own); it's reported once per
			// body as response-body-wrapped-in-one-of, so don't also report a body
			// type change.
			if !schemaDiff.OneOfWrappingDiff.Empty() {
				return
			}
			// Suppress null-only type changes (handled by nullable checkers).
			if isNullTypeChange(typeDiff) && formatDiff.Empty() {
				return
			}

			id := responseTypeChangeId(typeDiff, formatDiff, info.mediaType, schemaDiff,
				ResponseBodyTypeSpecializedId, ResponseBodyTypeGeneralizedId, ResponseBodyTypeChangedId)
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, schemaDiff, "type")
			result = append(result, info.newChange(
				id,
				[]any{getTypeFormatDimension(schemaDiff), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff), info.responseStatus},
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

			if !propTypeDiff.Empty() || !propFormatDiff.Empty() {
				id := responseTypeChangeId(propTypeDiff, propFormatDiff, info.mediaType, propSchemaDiff,
					ResponsePropertyTypeSpecializedId, ResponsePropertyTypeGeneralizedId, ResponsePropertyTypeChangedId)
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName), getTypeFormatDimension(propSchemaDiff), getBaseTypeFormat(propSchemaDiff), getRevisionTypeFormat(propSchemaDiff), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
