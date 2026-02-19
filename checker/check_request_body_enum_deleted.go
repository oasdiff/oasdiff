package checker

import (
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

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for _, mediaTypeItem := range mediaTypeChanges {
				if mediaTypeItem.SchemaDiff == nil {
					continue
				}

				enumDiff := mediaTypeItem.SchemaDiff.EnumDiff
				if enumDiff == nil || enumDiff.Deleted == nil {
					continue
				}
				for _, enumVal := range enumDiff.Deleted {
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
