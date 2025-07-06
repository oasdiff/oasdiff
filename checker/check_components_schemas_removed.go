package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APISchemasRemovedId = "api-schema-removed"
	ComponentSchemas    = "schemas"
)

func APIComponentsSchemaRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	if diffReport.ComponentsDiff == nil {
		return result
	}

	if diffReport.ComponentsDiff.SchemasDiff == nil {
		return result
	}

	for _, deletedSchema := range diffReport.ComponentsDiff.SchemasDiff.Deleted {
		result = append(result, ComponentChange{
			Id:        APISchemasRemovedId,
			Level:     config.getLogLevel(APISchemasRemovedId),
			Args:      []any{deletedSchema},
			Component: ComponentSchemas,
		})
	}
	return result
}
