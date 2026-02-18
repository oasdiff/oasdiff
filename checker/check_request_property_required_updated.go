package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyBecameRequiredId            = "request-property-became-required"
	RequestPropertyBecameRequiredWithDefaultId = "request-property-became-required-with-default"
	RequestPropertyBecameOptionalId            = "request-property-became-optional"
)

func RequestPropertyRequiredUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if mediaTypeDiff.SchemaDiff == nil {
					continue
				}

				processRequestPropertyRequiredDiff := func(schemaDiff *diff.SchemaDiff, propertyPath string, propertyName string) {
					if schemaDiff.RequiredDiff != nil {
						for _, changedRequiredPropertyName := range schemaDiff.RequiredDiff.Added {
							if !changedRequiredPropertyRelevant(schemaDiff, changedRequiredPropertyName) {
								continue
							}

							if schemaDiff.Revision.Properties[changedRequiredPropertyName].Value.Default == nil {
								result = append(result, NewApiChange(
									RequestPropertyBecameRequiredId,
									config,
									[]any{propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName))},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(baseSource, revisionSource))
							} else {
								// property has a default value, so making it required is not a breaking change
								result = append(result, NewApiChange(
									RequestPropertyBecameRequiredWithDefaultId,
									config,
									[]any{propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName))},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(baseSource, revisionSource))
							}
						}
						for _, changedRequiredPropertyName := range schemaDiff.RequiredDiff.Deleted {
							if !changedRequiredPropertyRelevant(schemaDiff, changedRequiredPropertyName) {
								continue
							}

							result = append(result, NewApiChange(
								RequestPropertyBecameOptionalId,
								config,
								[]any{propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName))},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						}
					}
				}

				processRequestPropertyRequiredDiff(mediaTypeDiff.SchemaDiff, "", "")

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, _ *diff.SchemaDiff) {
						processRequestPropertyRequiredDiff(propertyDiff, propertyPath, propertyName)
					})
			}
		}
	}
	return result
}

func changedRequiredPropertyRelevant(schemaDiff *diff.SchemaDiff, changedRequiredPropertyName string) bool {
	if schemaDiff.Base.Properties[changedRequiredPropertyName] == nil {
		// it is a new property, checked by the new-required-request-property check
		return false
	}
	if schemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
		// property was removed, checked by request-property-removed
		return false
	}
	if schemaDiff.Revision.Properties[changedRequiredPropertyName].Value.ReadOnly {
		// property is read-only, not relevant in requests
		return false
	}

	return true
}
