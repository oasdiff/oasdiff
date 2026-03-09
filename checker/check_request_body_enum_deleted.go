package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyEnumValueRemovedId = "request-body-enum-value-removed"
)

func RequestBodyEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil {
				continue
			}
			if operationItem.RequestBodyDiff.ContentDiff == nil {
				continue
			}
			if operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			mediaTypeChanges := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified

			for _, mediaTypeItem := range mediaTypeChanges {
				if mediaTypeItem.SchemaDiff == nil {
					continue
				}

				schemaDiff := mediaTypeItem.SchemaDiff
				enumDiff := schemaDiff.EnumDiff
				if enumDiff == nil || enumDiff.Deleted == nil {
					continue
				}
				for _, enumVal := range enumDiff.Deleted {
					baseSource, revisionSource := SchemaDeletedItemSources(operationsSources, operationItem, schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
					result = append(result, NewApiChange(
						RequestBodyEnumValueRemovedId,
						config,
						[]any{enumVal},
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
	return result
}
