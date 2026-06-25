package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APISecurityRemovedCheckId       = "api-security-removed"
	APISecurityAddedCheckId         = "api-security-added"
	APISecurityScopeAddedId         = "api-security-scope-added"
	APISecurityScopeRemovedId       = "api-security-scope-removed"
	APIGlobalSecurityRemovedCheckId = "api-global-security-removed"
	APIGlobalSecurityAddedCheckId   = "api-global-security-added"
	APIGlobalSecurityScopeAddedId   = "api-global-security-scope-added"
	APIGlobalSecurityScopeRemovedId = "api-global-security-scope-removed"
)

func checkGlobalSecurity(diffReport *diff.Diff) Changes {
	result := make(Changes, 0)
	if diffReport.SecurityDiff == nil {
		return result
	}

	for _, addedSecurity := range diffReport.SecurityDiff.Added {
		result = append(result, SecurityChange{
			Id:    APIGlobalSecurityAddedCheckId,
			Level: INFO,
			Args:  []any{addedSecurity.String()},
		})
	}

	for _, removedSecurity := range diffReport.SecurityDiff.Deleted {
		result = append(result, SecurityChange{
			Id:    APIGlobalSecurityRemovedCheckId,
			Level: INFO,
			Args:  []any{removedSecurity.String()},
		})
	}

	for _, updatedSecurity := range diffReport.SecurityDiff.Modified {
		for securitySchemeName, updatedSecuritySchemeScopes := range updatedSecurity.Scopes {
			for _, addedScope := range updatedSecuritySchemeScopes.Added {
				result = append(result, SecurityChange{
					Id:    APIGlobalSecurityScopeAddedId,
					Level: INFO,
					Args:  []any{addedScope, securitySchemeName},
				})
			}
			for _, deletedScope := range updatedSecuritySchemeScopes.Deleted {
				result = append(result, SecurityChange{
					Id:    APIGlobalSecurityScopeRemovedId,
					Level: INFO,
					Args:  []any{deletedScope, securitySchemeName},
				})
			}
		}
	}

	return result
}

func APISecurityUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	result = append(result, checkGlobalSecurity(diffReport)...)

	if diffReport.PathsDiff == nil || diffReport.PathsDiff.Modified == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {

			if operationItem.SecurityDiff == nil {
				continue
			}

			for _, addedSecurity := range operationItem.SecurityDiff.Added {
				if len(addedSecurity.Schemes) == 0 {
					continue
				}

				result = append(result, NewApiChange(
					APISecurityAddedCheckId,
					config,
					[]any{addedSecurity.String()},
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				))
			}

			for _, deletedSecurity := range operationItem.SecurityDiff.Deleted {
				if len(deletedSecurity.Schemes) == 0 {
					continue
				}

				result = append(result, NewApiChange(
					APISecurityRemovedCheckId,
					config,
					[]any{deletedSecurity.String()},
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				))
			}

			for _, updatedSecurity := range operationItem.SecurityDiff.Modified {
				if updatedSecurity.Scopes.Empty() {
					continue
				}
				for securitySchemeName, updatedSecuritySchemeScopes := range updatedSecurity.Scopes {
					for _, addedScope := range updatedSecuritySchemeScopes.Added {
						result = append(result, NewApiChange(
							APISecurityScopeAddedId,
							config,
							[]any{addedScope, securitySchemeName},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}
					for _, deletedScope := range updatedSecuritySchemeScopes.Deleted {
						result = append(result, NewApiChange(
							APISecurityScopeRemovedId,
							config,
							[]any{deletedScope, securitySchemeName},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}
				}
			}
		}
	}

	return result
}
