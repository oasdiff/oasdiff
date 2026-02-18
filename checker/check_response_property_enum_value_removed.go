package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyEnumValueRemovedId = "response-property-enum-value-removed"
)

func ResponseParameterEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff == nil ||
					responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							enumDiff := propertyDiff.EnumDiff
							if enumDiff == nil || enumDiff.Deleted == nil {
								return
							}

							for _, enumVal := range enumDiff.Deleted {
								result = append(result, NewApiChange(
									ResponsePropertyEnumValueRemovedId,
									config,
									[]any{enumVal, propertyFullName(propertyPath, propertyName), responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
							}
						})
				}
			}
		}
	}
	return result
}
