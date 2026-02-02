package checker

import (
	"github.com/oasdiff/oasdiff/diff"
	"slices"
)

const (
	ResponseOptionalPropertyBecameNonWriteOnlyId = "response-optional-property-became-not-write-only"
	ResponseOptionalPropertyBecameWriteOnlyId    = "response-optional-property-became-write-only"
	ResponseOptionalPropertyBecameReadOnlyId     = "response-optional-property-became-read-only"
	ResponseOptionalPropertyBecameNonReadOnlyId  = "response-optional-property-became-not-read-only"
)

func ResponseOptionalPropertyWriteOnlyReadOnlyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for _, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							writeOnlyDiff := propertyDiff.WriteOnlyDiff
							if writeOnlyDiff == nil {
								return
							}
							if parent.Revision.Properties[propertyName] == nil {
								// removed properties processed by the ResponseOptionalPropertyUpdatedCheck check
								return
							}
							if slices.Contains(parent.Base.Required, propertyName) {
								// skip required properties - checked at ResponseRequiredPropertyWriteOnlyReadOnlyCheck
								return
							}

							id := ResponseOptionalPropertyBecameNonWriteOnlyId

							if writeOnlyDiff.To == true {
								id = ResponseOptionalPropertyBecameWriteOnlyId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							))
						})

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							readOnlyDiff := propertyDiff.ReadOnlyDiff
							if readOnlyDiff == nil {
								return
							}
							if parent.Revision.Properties[propertyName] == nil {
								// removed properties processed by the ResponseOptionalPropertyUpdatedCheck check
								return
							}
							if slices.Contains(parent.Base.Required, propertyName) {
								// skip non-optional properties
								return
							}

							id := ResponseOptionalPropertyBecameNonReadOnlyId

							if readOnlyDiff.To == true {
								id = ResponseOptionalPropertyBecameReadOnlyId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							))
						})
				}
			}
		}
	}
	return result
}
