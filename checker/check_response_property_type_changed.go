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
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff != nil {
						// Check for suppression by ListOfTypes checker
						if shouldSuppressTypeChangedForListOfTypes(mediaTypeDiff.SchemaDiff) {
							continue
						}

						schemaDiff := mediaTypeDiff.SchemaDiff
						typeDiff := schemaDiff.TypeDiff
						formatDiff := schemaDiff.FormatDiff
						if breakingTypeFormatChangedInResponseProperty(typeDiff, formatDiff, mediaType, schemaDiff) {

							result = append(result, NewApiChange(
								ResponseBodyTypeChangedId,
								config,
								[]any{getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							))
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff == nil || propertyDiff.Revision == nil {
								return
							}

							// Check for suppression by ListOfTypes checker
							if shouldSuppressPropertyTypeChangedForListOfTypes(propertyDiff) {
								return
							}

							schemaDiff := propertyDiff
							typeDiff := schemaDiff.TypeDiff
							formatDiff := schemaDiff.FormatDiff

							if breakingTypeFormatChangedInResponseProperty(typeDiff, formatDiff, mediaType, schemaDiff) {

								result = append(result, NewApiChange(
									ResponsePropertyTypeChangedId,
									config,
									[]any{propertyFullName(propertyPath, propertyName), getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff), responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								))
							}
						})
				}
			}
		}
	}
	return result
}
