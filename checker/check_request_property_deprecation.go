package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyDeprecatedId              = "request-property-deprecated"
	RequestPropertyDeprecatedSunsetMissingId = "request-property-deprecated-sunset-missing"
	RequestPropertyDeprecatedInvalidId       = "request-property-deprecated-sunset-invalid"
)

// RequestPropertyDeprecationCheck detects deprecated properties in request bodies
func RequestPropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			// Track reported properties to avoid duplicates per operation
			reportedProperties := make(map[string]bool)

			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for _, mediaTypeDiff := range modifiedMediaTypes {
				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						// Check if property was newly deprecated
						if propertyDiff.DeprecatedDiff == nil {
							return
						}
						if propertyDiff.DeprecatedDiff.To != true {
							return
						}
						if propertyDiff.DeprecatedDiff.From != nil && propertyDiff.DeprecatedDiff.From != false {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						// Skip if already reported
						if reportedProperties[propName] {
							return
						}
						reportedProperties[propName] = true

						changeId, args := getRequestPropertyDeprecationId(propName, propertyDiff)
						result = append(result, NewApiChange(
							changeId,
							config,
							args,
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
	return result
}

func getRequestPropertyDeprecationId(propName string, propertyDiff *diff.SchemaDiff) (string, []any) {
	if propertyDiff == nil || propertyDiff.Revision == nil {
		return RequestPropertyDeprecatedSunsetMissingId, []any{propName}
	}
	sunset, ok := getSunset(propertyDiff.Revision.Extensions)
	if !ok {
		return RequestPropertyDeprecatedSunsetMissingId, []any{propName}
	}
	date, err := getSunsetDate(sunset)
	if err != nil {
		return RequestPropertyDeprecatedInvalidId, []any{propName, err}
	}
	return RequestPropertyDeprecatedId, []any{propName, date}
}
