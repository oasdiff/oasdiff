package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tufin/oasdiff/diff"
)

func ResponseSuccessStatusUpdated(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config Config) Changes {
	success := func(status int) bool {
		return status >= 200 && status <= 299
	}

	return ResponseStatusUpdated(diffReport, operationsSources, config, success, "response-success-status-removed", ERR)
}

func ResponseNonSuccessStatusUpdated(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config Config) Changes {
	notSuccess := func(status int) bool {
		return status < 200 || status > 299
	}

	return ResponseStatusUpdated(diffReport, operationsSources, config, notSuccess, "response-non-success-status-removed", INFO)
}

func ResponseStatusUpdated(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config Config, filter func(int) bool, id string, defaultLevel Level) Changes {
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
			source := (*operationsSources)[operationItem.Revision]
			for _, responseStatus := range operationItem.ResponsesDiff.Deleted {
				status, err := strconv.Atoi(responseStatus)
				if err != nil {
					continue
				}

				if filter(status) {
					result = append(result, ApiChange{
						Id:          id,
						Level:       config.getLogLevel(id, defaultLevel),
						Text:        fmt.Sprintf(config.i18n(id), colorizedValue(responseStatus)),
						Operation:   operation,
						OperationId: operationItem.Revision.OperationID,
						Path:        path,
						Source:      source,
					})
				}

			}

			for _, responseStatus := range operationItem.ResponsesDiff.Added {
				addedId := strings.Replace(id, "removed", "added", 1)
				defaultLevel := INFO
				status, err := strconv.Atoi(responseStatus)
				if err != nil {
					continue
				}

				if filter(status) {
					result = append(result, ApiChange{
						Id:          addedId,
						Level:       config.getLogLevel(addedId, defaultLevel),
						Text:        fmt.Sprintf(config.i18n(addedId), colorizedValue(responseStatus)),
						Operation:   operation,
						OperationId: operationItem.Revision.OperationID,
						Path:        path,
						Source:      source,
					})
				}

			}

		}
	}
	return result
}
