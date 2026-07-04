package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyTypeChangedId         = "response-body-type-changed"
	ResponseBodyTypeGeneralizedId     = "response-body-type-generalized"
	ResponseBodyTypeSpecializedId     = "response-body-type-specialized"
	ResponseBodyTypeCompatibleId      = "response-body-type-compatible"
	ResponsePropertyTypeChangedId     = "response-property-type-changed"
	ResponsePropertyTypeGeneralizedId = "response-property-type-generalized"
	ResponsePropertyTypeSpecializedId = "response-property-type-specialized"
	ResponsePropertyTypeCompatibleId  = "response-property-type-compatible"
)

// responseTypeChangeId classifies a response type/format change (the
// contravariant mirror of the request side):
//   - specialized (info): the returned type set narrowed (number -> integer).
//   - compatible (info): a type swap that's safe only because the media type
//     isn't strongly typed (string -> object in application/xml), not a narrowing.
//   - generalized (error): the returned type set widened (integer -> number).
//   - changed (error): a genuinely incompatible change (string -> integer).
func responseTypeChangeId(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff, specializedId, compatibleId, generalizedId, changedId string) (id, comment string) {
	if responseTypeFormatBreaking(typeDiff, formatDiff, mediaType, schemaDiff) {
		if typeSetWidened(typeDiff, schemaDiff) {
			return generalizedId, ""
		}
		return changedId, ""
	}
	if isResponseLooseTypeSwap(typeDiff, schemaDiff) {
		return compatibleId, TypeChangeLooselyTypedCommentId
	}
	return specializedId, ""
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

			id, comment := responseTypeChangeId(typeDiff, formatDiff, info.mediaType, schemaDiff,
				ResponseBodyTypeSpecializedId, ResponseBodyTypeCompatibleId, ResponseBodyTypeGeneralizedId, ResponseBodyTypeChangedId)
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, schemaDiff, "type")
			result = append(result, info.newChange(
				id,
				[]any{getTypeFormatDimension(schemaDiff), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff), info.responseStatus},
				comment,
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
				id, comment := responseTypeChangeId(propTypeDiff, propFormatDiff, info.mediaType, propSchemaDiff,
					ResponsePropertyTypeSpecializedId, ResponsePropertyTypeCompatibleId, ResponsePropertyTypeGeneralizedId, ResponsePropertyTypeChangedId)
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName), getTypeFormatDimension(propSchemaDiff), getBaseTypeFormat(propSchemaDiff), getRevisionTypeFormat(propSchemaDiff), info.responseStatus},
					comment,
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
