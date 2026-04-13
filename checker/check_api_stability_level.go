package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyStabilityDecreasedId  = "request-property-stability-decreased"
	RequestPropertyStabilityIncreasedId  = "request-property-stability-increased"
	ResponsePropertyStabilityDecreasedId = "response-property-stability-decreased"
	ResponsePropertyStabilityIncreasedId = "response-property-stability-increased"
)

// RequestPropertyStabilityUpdatedCheck detects request properties where x-stability-level changed.
// Only emits changes when the property's base stability meets the configured threshold.
func RequestPropertyStabilityUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if config == nil || diffReport.PathsDiff == nil {
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
			op := operationItem.Revision
			for _, mediaTypeDiff := range operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified {
				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						checkPropertyStabilityChange(propertyDiff, propertyPath, propertyName,
							RequestPropertyStabilityDecreasedId, RequestPropertyStabilityIncreasedId,
							config, operationsSources, op, operation, path, &result)
					})
			}
		}
	}
	return result
}

// ResponsePropertyStabilityUpdatedCheck detects response properties where x-stability-level changed.
// Only emits changes when the property's base stability meets the configured threshold.
func ResponsePropertyStabilityUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if config == nil || diffReport.PathsDiff == nil {
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
			op := operationItem.Revision
			for _, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff == nil ||
					responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				for _, mediaTypeDiff := range responseDiff.ContentDiff.MediaTypeModified {
					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							checkPropertyStabilityChange(propertyDiff, propertyPath, propertyName,
								ResponsePropertyStabilityDecreasedId, ResponsePropertyStabilityIncreasedId,
								config, operationsSources, op, operation, path, &result)
						})
				}
			}
		}
	}
	return result
}

func checkPropertyStabilityChange(
	propertyDiff *diff.SchemaDiff,
	propertyPath string, propertyName string,
	decreasedId string, increasedId string,
	config *Config, operationsSources *diff.OperationsSourcesMap,
	op *openapi3.Operation, operation string, path string,
	result *Changes,
) {
	if propertyDiff.ExtensionsDiff == nil {
		return
	}

	var baseStability, revisionStability string
	var err error

	if propertyDiff.Base != nil {
		baseStability, err = getStabilityLevel(propertyDiff.Base.Extensions)
		if err != nil {
			return
		}
	}
	if propertyDiff.Revision != nil {
		revisionStability, err = getStabilityLevel(propertyDiff.Revision.Extensions)
		if err != nil {
			return
		}
	}

	// Normalize: no x-stability-level is treated as implicit "stable"
	baseLabel := baseStability
	if baseLabel == "" {
		baseLabel = STABILITY_STABLE
	}
	revisionLabel := revisionStability
	if revisionLabel == "" {
		revisionLabel = STABILITY_STABLE
	}

	baseLevel := ParseStabilityLevel(baseLabel)
	revisionLevel := ParseStabilityLevel(revisionLabel)

	if baseLevel == revisionLevel {
		return
	}

	// Report only if both base and revision meet the configured threshold
	if !config.StabilityLevel.IsIncluded(lowerStability(baseLabel, revisionLabel)) {
		return
	}

	isDecrease := revisionLevel < baseLevel

	propName := propertyFullName(propertyPath, propertyName)

	var changeId string
	if isDecrease {
		changeId = decreasedId
	} else {
		changeId = increasedId
	}

	*result = append(*result, NewApiChange(
		changeId,
		config,
		[]any{propName, baseLabel, revisionLabel},
		"",
		operationsSources,
		op,
		operation,
		path,
	))
}
