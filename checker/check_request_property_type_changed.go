package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyTypeGeneralizedId     = "request-body-type-generalized"
	RequestBodyTypeChangedId         = "request-body-type-changed"
	RequestBodyTypeCompatibleId      = "request-body-type-compatible"
	RequestPropertyTypeGeneralizedId = "request-property-type-generalized"
	RequestPropertyTypeChangedId     = "request-property-type-changed"
	RequestPropertyTypeCompatibleId  = "request-property-type-compatible"
)

// requestTypeChangeId classifies a request type/format change:
//   - changed (error): breaking.
//   - compatible (info): a type swap safe only because the media type isn't
//     strongly typed (object -> string in application/xml), not a real widening.
//   - generalized (info): otherwise (a genuine widening, or a safe format change).
func requestTypeChangeId(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff, generalizedId, compatibleId, changedId string) (id, comment string) {
	if requestTypeFormatBreaking(typeDiff, formatDiff, mediaType, schemaDiff) {
		return changedId, ""
	}
	if isRequestLooseTypeSwap(typeDiff, schemaDiff) {
		return compatibleId, TypeChangeLooselyTypedCommentId
	}
	return generalizedId, ""
}

func RequestPropertyTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		schemaDiff := info.schemaDiff
		typeDiff := schemaDiff.TypeDiff
		formatDiff := schemaDiff.FormatDiff

		if !typeDiff.Empty() || !formatDiff.Empty() {
			id, comment := requestTypeChangeId(typeDiff, formatDiff, info.mediaType, schemaDiff,
				RequestBodyTypeGeneralizedId, RequestBodyTypeCompatibleId, RequestBodyTypeChangedId)
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, schemaDiff, "type")
			result = append(result, info.newChange(
				id,
				[]any{getTypeFormatDimension(schemaDiff), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff)},
				comment,
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.Revision == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
				return
			}

			propSchemaDiff := p.propertyDiff
			propTypeDiff := propSchemaDiff.TypeDiff
			propFormatDiff := propSchemaDiff.FormatDiff

			if !propTypeDiff.Empty() || !propFormatDiff.Empty() {
				id, comment := requestTypeChangeId(propTypeDiff, propFormatDiff, info.mediaType, propSchemaDiff,
					RequestPropertyTypeGeneralizedId, RequestPropertyTypeCompatibleId, RequestPropertyTypeChangedId)
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName), getTypeFormatDimension(propSchemaDiff), getBaseTypeFormat(propSchemaDiff), getRevisionTypeFormat(propSchemaDiff)},
					comment,
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
