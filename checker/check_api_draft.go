package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	EndpointDraftId         = "endpoint-draft"
	RequestPropertyDraftId  = "request-property-draft"
	ResponsePropertyDraftId = "response-property-draft"
)

// APIDraftCheck detects endpoints where x-stability-level is set to "draft" in the revision
// but was not "draft" in the base spec.
// Only emits changes when IncludeStabilityLevels includes "draft".
func APIDraftCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if config == nil || !config.IncludeStabilityLevels[STABILITY_DRAFT] {
		return result
	}
	if diffReport.PathsDiff == nil {
		return result
	}

	// Check newly added paths
	for _, path := range diffReport.PathsDiff.Added {
		for opName, op := range diffReport.PathsDiff.Revision.Value(path).Operations() {
			stability, err := getStabilityLevel(op.Extensions)
			if err != nil || stability != STABILITY_DRAFT {
				continue
			}
			result = append(result, NewApiChange(
				EndpointDraftId,
				config,
				nil,
				"",
				operationsSources,
				op,
				opName,
				path,
			))
		}
	}

	// Check modified paths for newly draft operations
	for path, pathDiff := range diffReport.PathsDiff.Modified {
		if pathDiff.OperationsDiff == nil {
			continue
		}
		for operation, operationDiff := range pathDiff.OperationsDiff.Modified {
			revisionOp := pathDiff.Revision.GetOperation(operation)
			if revisionOp == nil {
				continue
			}
			revisionStability, err := getStabilityLevel(revisionOp.Extensions)
			if err != nil || revisionStability != STABILITY_DRAFT {
				continue
			}
			// Check if it was already draft in base
			baseOp := pathDiff.Base.GetOperation(operation)
			if baseOp != nil {
				baseStability, err := getStabilityLevel(baseOp.Extensions)
				if err == nil && baseStability == STABILITY_DRAFT {
					continue // was already draft, not a new change
				}
			}
			result = append(result, NewApiChange(
				EndpointDraftId,
				config,
				nil,
				"",
				operationsSources,
				operationDiff.Revision,
				operation,
				path,
			))
		}
	}

	return result
}

// RequestPropertyDraftCheck detects request properties where x-stability-level is set to "draft"
// in the revision but was not in the base spec.
// Only emits changes when IncludeStabilityLevels includes "draft".
func RequestPropertyDraftCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if config == nil || !config.IncludeStabilityLevels[STABILITY_DRAFT] {
		return result
	}
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

			op := operationItem.Revision

			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for _, mediaTypeDiff := range modifiedMediaTypes {
				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff.ExtensionsDiff == nil {
							return
						}
						propName := propertyFullName(propertyPath, propertyName)

						// Check if x-stability-level was added or changed to "draft" in revision
						revStability, err := getStabilityLevel(propertyDiff.Revision.Extensions)
						if err != nil || revStability != STABILITY_DRAFT {
							return
						}
						// Check if it was already draft in base
						if propertyDiff.Base != nil {
							baseStability, err := getStabilityLevel(propertyDiff.Base.Extensions)
							if err == nil && baseStability == STABILITY_DRAFT {
								return // was already draft
							}
						}
						result = append(result, NewApiChange(
							RequestPropertyDraftId,
							config,
							[]any{propName},
							"",
							operationsSources,
							op,
							operation,
							path,
						))
					})
			}
		}
	}
	return result
}

// ResponsePropertyDraftCheck detects response properties where x-stability-level is set to "draft"
// in the revision but was not in the base spec.
// Only emits changes when IncludeStabilityLevels includes "draft".
func ResponsePropertyDraftCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if config == nil || !config.IncludeStabilityLevels[STABILITY_DRAFT] {
		return result
	}
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

			op := operationItem.Revision

			for _, responseDiff := range operationItem.ResponsesDiff.Modified {
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
							if propertyDiff.ExtensionsDiff == nil {
								return
							}
							propName := propertyFullName(propertyPath, propertyName)

							// Check if x-stability-level was added or changed to "draft" in revision
							revStability, err := getStabilityLevel(propertyDiff.Revision.Extensions)
							if err != nil || revStability != STABILITY_DRAFT {
								return
							}
							// Check if it was already draft in base
							if propertyDiff.Base != nil {
								baseStability, err := getStabilityLevel(propertyDiff.Base.Extensions)
								if err == nil && baseStability == STABILITY_DRAFT {
									return // was already draft
								}
							}
							result = append(result, NewApiChange(
								ResponsePropertyDraftId,
								config,
								[]any{propName},
								"",
								operationsSources,
								op,
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
