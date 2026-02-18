package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseMediaTypeEnumValueRemovedId = "response-mediatype-enum-value-removed"
)

func ResponseMediaTypeEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil {
				continue
			}
			if operationItem.ResponsesDiff.Modified == nil {
				continue
			}
			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for _, responseItems := range operationItem.ResponsesDiff.Modified {
				if responseItems.ContentDiff == nil {
					continue
				}

				if responseItems.ContentDiff.MediaTypeModified == nil {
					continue
				}
				for mediaType, mediaTypeItem := range responseItems.ContentDiff.MediaTypeModified {
					if mediaTypeItem.SchemaDiff == nil {
						continue
					}

					enumDiff := mediaTypeItem.SchemaDiff.EnumDiff
					if enumDiff == nil {
						continue
					}

					for _, enumVal := range enumDiff.Deleted {
						result = append(result, NewApiChange(
							ResponseMediaTypeEnumValueRemovedId,
							config,
							[]any{mediaType, enumVal},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}
				}
			}
		}
	}
	return result
}
