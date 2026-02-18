package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyEnumValueRemovedId         = "request-property-enum-value-removed"
	RequestReadOnlyPropertyEnumValueRemovedId = "request-read-only-property-enum-value-removed"
	RequestPropertyEnumValueAddedId           = "request-property-enum-value-added"
)

func RequestPropertyEnumValueUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for _, mediaTypeDiff := range modifiedMediaTypes {
				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						enumDiff := propertyDiff.EnumDiff
						if enumDiff == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						for _, enumVal := range enumDiff.Deleted {

							id := RequestPropertyEnumValueRemovedId
							if propertyDiff.Revision.ReadOnly {
								id = RequestReadOnlyPropertyEnumValueRemovedId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{enumVal, propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						}

						for _, enumVal := range enumDiff.Added {
							result = append(result, NewApiChange(
								RequestPropertyEnumValueAddedId,
								config,
								[]any{enumVal, propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						}
					})
			}
		}
	}
	return result
}
