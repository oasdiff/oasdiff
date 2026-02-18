package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	NewRequiredRequestDefaultParameterToExistingPathId = "new-required-request-default-parameter-to-existing-path"
	NewOptionalRequestDefaultParameterToExistingPathId = "new-optional-request-default-parameter-to-existing-path"
)

func NewRequestNonPathDefaultParameterCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil || len(diffReport.PathsDiff.Modified) == 0 {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.ParametersDiff == nil || pathItem.Revision == nil || len(pathItem.Revision.Operations()) == 0 {
			continue
		}

		for paramLoc, paramNameList := range pathItem.ParametersDiff.Added {
			if paramLoc == "path" {
				continue
			}

			for _, param := range pathItem.Revision.Parameters {
				if !slices.Contains(paramNameList, param.Value.Name) {
					continue
				}
				id := NewRequiredRequestDefaultParameterToExistingPathId
				if !param.Value.Required {
					id = NewOptionalRequestDefaultParameterToExistingPathId
				}

				for operation, operationItem := range pathItem.Revision.Operations() {

					// TODO: if base operation had this param individually (not through the path) - continue

					var baseSource, revisionSource *Source
					baseOp := pathItem.Base.GetOperation(operation)
					hasOrigin := (baseOp != nil && baseOp.Origin != nil) || operationItem.Origin != nil
					if hasOrigin {
						if baseOp != nil {
							baseSource = NewSourceFromOrigin(operationsSources, baseOp, baseOp.Origin)
						}
						revisionSource = NewSourceFromOrigin(operationsSources, operationItem, operationItem.Origin)
					}

					result = append(result, NewApiChange(
						id,
						config,
						[]any{paramLoc, param.Value.Name},
						"",
						operationsSources,
						operationItem,
						operation,
						path,
					).WithSources(baseSource, revisionSource))
				}
			}
		}
	}
	return result
}
