package checker

import (
	"fmt"

	"github.com/tufin/oasdiff/diff"
)

const (
	ResponsePropertyPatternAddedId   = "response-property-pattern-added"
	ResponsePropertyPatternChangedId = "response-property-pattern-changed"
	ResponsePropertyPatternRemovedId = "response-property-pattern-removed"
)

func ResponsePatternAddedOrChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			source := (*operationsSources)[operationItem.Revision]

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

					checkModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							patternDiff := propertyDiff.PatternDiff
							if patternDiff == nil {
								return
							}

							id := ResponsePropertyPatternChangedId
							text := fmt.Sprintf(config.i18n(id), colorizedValue(propertyFullName(propertyPath, propertyName)), colorizedValue(patternDiff.From), colorizedValue(patternDiff.To), colorizedValue(responseStatus))
							if patternDiff.To == "" || patternDiff.To == nil {
								id = ResponsePropertyPatternRemovedId
								text = fmt.Sprintf(config.i18n(id), colorizedValue(propertyFullName(propertyPath, propertyName)), colorizedValue(patternDiff.From), colorizedValue(responseStatus))

							} else if patternDiff.From == "" || patternDiff.From == nil {
								id = ResponsePropertyPatternAddedId
								text = fmt.Sprintf(config.i18n(id), colorizedValue(propertyFullName(propertyPath, propertyName)), colorizedValue(patternDiff.To), colorizedValue(responseStatus))
							}

							result = append(result, ApiChange{
								Id:          id,
								Level:       INFO,
								Text:        text,
								Operation:   operation,
								OperationId: operationItem.Revision.OperationID,
								Path:        path,
								Source:      source,
							})

						})
				}
			}
		}
	}
	return result
}
