package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyDeprecatedId              = "response-property-deprecated"
	ResponsePropertyDeprecatedSunsetMissingId = "response-property-deprecated-sunset-missing"
	ResponsePropertyDeprecatedInvalidId       = "response-property-deprecated-sunset-invalid"
)

// ResponsePropertyDeprecationCheck detects deprecated properties in response bodies
func ResponsePropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			// Track reported properties to avoid duplicates per operation
			reportedProperties := make(map[string]bool)

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff == nil ||
					responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
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

							changeId, args := getResponsePropertyDeprecationId(propName, propertyDiff, responseStatus)
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
	}
	return result
}

func getResponsePropertyDeprecationId(propName string, propertyDiff *diff.SchemaDiff, responseStatus string) (string, []any) {
	if propertyDiff == nil || propertyDiff.Revision == nil {
		return ResponsePropertyDeprecatedSunsetMissingId, []any{propName, responseStatus}
	}
	sunset, ok := getSunset(propertyDiff.Revision.Extensions)
	if !ok {
		return ResponsePropertyDeprecatedSunsetMissingId, []any{propName, responseStatus}
	}
	date, err := getSunsetDate(sunset)
	if err != nil {
		return ResponsePropertyDeprecatedInvalidId, []any{propName, responseStatus, err}
	}
	return ResponsePropertyDeprecatedId, []any{propName, responseStatus, date}
}
