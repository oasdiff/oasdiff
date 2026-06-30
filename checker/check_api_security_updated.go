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

	// The document-root "security" field location in each spec; nil when origin
	// tracking is off. Added/scope-added are reported against the revision, the
	// rest against the base, matching the add/remove source convention.
	baseSource := sourceFromField(diffReport.SecurityDiff.BaseOrigin, "security")
	revisionSource := sourceFromField(diffReport.SecurityDiff.RevisionOrigin, "security")

	for _, addedSecurity := range diffReport.SecurityDiff.Added {
		result = append(result, SecurityChange{
			Id:    APIGlobalSecurityAddedCheckId,
			Level: INFO,
			Args:  []any{addedSecurity.String()},
		}.WithSources(nil, revisionSource))
	}

	for _, removedSecurity := range diffReport.SecurityDiff.Deleted {
		result = append(result, SecurityChange{
			Id:    APIGlobalSecurityRemovedCheckId,
			Level: INFO,
			Args:  []any{removedSecurity.String()},
		}.WithSources(baseSource, nil))
	}

	for _, updatedSecurity := range diffReport.SecurityDiff.Modified {
		for securitySchemeName, updatedSecuritySchemeScopes := range updatedSecurity.Scopes {
			for _, addedScope := range updatedSecuritySchemeScopes.Added {
				result = append(result, SecurityChange{
					Id:    APIGlobalSecurityScopeAddedId,
					Level: INFO,
					Args:  []any{addedScope, securitySchemeName},
				}.WithSources(nil, revisionSource))
			}
			for _, deletedScope := range updatedSecuritySchemeScopes.Deleted {
				result = append(result, SecurityChange{
					Id:    APIGlobalSecurityScopeRemovedId,
					Level: INFO,
					Args:  []any{deletedScope, securitySchemeName},
				}.WithSources(baseSource, nil))
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

			baseSource := securitySource(operationsSources, operationItem.Base)
			revisionSource := securitySource(operationsSources, operationItem.Revision)

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)

			for _, addedSecurity := range operationItem.SecurityDiff.Added {
				if len(addedSecurity.Schemes) == 0 {
					continue
				}

				result = append(result, opInfo.NewApiChange(
					APISecurityAddedCheckId,
					[]any{addedSecurity.String()},
					"",
				).WithSources(nil, revisionSource))
			}

			for _, deletedSecurity := range operationItem.SecurityDiff.Deleted {
				if len(deletedSecurity.Schemes) == 0 {
					continue
				}

				result = append(result, opInfo.NewApiChange(
					APISecurityRemovedCheckId,
					[]any{deletedSecurity.String()},
					"",
				).WithSources(baseSource, nil))
			}

			for _, updatedSecurity := range operationItem.SecurityDiff.Modified {
				if updatedSecurity.Scopes.Empty() {
					continue
				}
				for securitySchemeName, updatedSecuritySchemeScopes := range updatedSecurity.Scopes {
					for _, addedScope := range updatedSecuritySchemeScopes.Added {
						result = append(result, opInfo.NewApiChange(
							APISecurityScopeAddedId,
							[]any{addedScope, securitySchemeName},
							"",
						).WithSources(nil, revisionSource))
					}
					for _, deletedScope := range updatedSecuritySchemeScopes.Deleted {
						result = append(result, opInfo.NewApiChange(
							APISecurityScopeRemovedId,
							[]any{deletedScope, securitySchemeName},
							"",
						).WithSources(baseSource, nil))
					}
				}
			}
		}
	}

	return result
}
