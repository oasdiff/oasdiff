package checker

import (
	"fmt"

	"github.com/tufin/oasdiff/diff"
)

func RequestParameterEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ParametersDiff == nil {
				continue
			}
			if operationItem.ParametersDiff.Modified == nil {
				continue
			}
			source := (*operationsSources)[operationItem.Revision]
			for paramLocation, paramItems := range operationItem.ParametersDiff.Modified {
				for paramName, paramItem := range paramItems {
					if paramItem.SchemaDiff == nil {
						continue
					}
					enumDiff := paramItem.SchemaDiff.EnumDiff
					if enumDiff == nil || enumDiff.Deleted == nil {
						continue
					}
					for _, enumVal := range enumDiff.Deleted {
						result = append(result, ApiChange{
							Id:          "request-parameter-enum-value-removed",
							Level:       ERR,
							Text:        fmt.Sprintf(config.i18n("request-parameter-enum-value-removed"), enumVal, colorizedValue(paramLocation), colorizedValue(paramName)),
							Operation:   operation,
							OperationId: operationItem.Revision.OperationID,
							Path:        path,
							Source:      source,
						})
					}
				}
			}
		}
	}
	return result
}
