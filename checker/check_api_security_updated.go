package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/location"
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

	APISecurityAnonymousAccessRemovedId       = "api-security-anonymous-access-removed"
	APISecurityAnonymousAccessAddedId         = "api-security-anonymous-access-added"
	APIGlobalSecurityAnonymousAccessRemovedId = "api-global-security-anonymous-access-removed"
	APIGlobalSecurityAnonymousAccessAddedId   = "api-global-security-anonymous-access-added"
)

func checkGlobalSecurity(diffReport *diff.Diff, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.SecurityDiff == nil {
		return result
	}

	// The document-root "security" field location in each spec; nil when origin
	// tracking is off. Added/scope-added are reported against the revision, the
	// rest against the base, matching the add/remove source convention.
	baseSource := location.SourceFromField(diffReport.SecurityDiff.BaseOrigin, "security")
	revisionSource := location.SourceFromField(diffReport.SecurityDiff.RevisionOrigin, "security")

	if securityContext := diffReport.SecurityContext; securityContext != nil && inheritsRootSecurity(securityContext.Operations) {
		access := effectiveAnonymousAccess(&securityContext.Base, &securityContext.Revision)
		if access.changed() {
			return Changes{globalAnonymousAccessChange(config, access, securityContext.Revision).WithSources(baseSource, revisionSource)}
		}
	}

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

	result = append(result, checkGlobalSecurity(diffReport, config)...)

	inherited, handled := checkInheritedSecurity(diffReport, operationsSources, config)
	result = append(result, inherited...)

	if diffReport.PathsDiff == nil || diffReport.PathsDiff.Modified == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {

			if operationItem.SecurityDiff == nil || handled[diff.Endpoint{Method: operation, Path: path}] {
				continue
			}

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			result = append(result, operationSecurityChanges(opInfo, operationsSources, operationItem.Base, operationItem.Revision, operationItem.Revision.Security, operationItem.SecurityDiff, declaredAnonymousAccess(operationItem))...)
		}
	}

	return result
}

// checkInheritedSecurity reports the operations whose security comes, on at
// least one side, from the document-root list. It returns the endpoints it
// reported on, so the caller skips their own security diff, which does not
// account for the inherited list.
func checkInheritedSecurity(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) (Changes, map[diff.Endpoint]bool) {
	result := make(Changes, 0)
	handled := map[diff.Endpoint]bool{}

	securityContext := diffReport.SecurityContext
	if securityContext == nil {
		return result, handled
	}

	for _, pair := range securityContext.Operations {
		if pair.Base.Security != nil && pair.Revision.Security != nil {
			continue
		}

		if pair.Base.Security == nil && pair.Revision.Security == nil {
			// Only the root list changed for this operation, so the change is
			// reported once, against the root list, rather than per operation.
			continue
		}

		baseSecurity := effectiveSecurity(pair.Base, &securityContext.Base)
		revisionSecurity := effectiveSecurity(pair.Revision, &securityContext.Revision)
		access := effectiveAnonymousAccess(baseSecurity, revisionSecurity)

		handled[diff.Endpoint{Method: pair.Method, Path: pair.Path}] = true
		opInfo := newOpInfo(config, pair.Revision, operationsSources, pair.Method, pair.Path)
		securityDiff := diff.GetSecurityRequirementsDiff(baseSecurity, revisionSecurity)
		result = append(result, operationSecurityChanges(opInfo, operationsSources, pair.Base, pair.Revision, revisionSecurity, securityDiff, access)...)
	}

	return result, handled
}

func inheritsRootSecurity(operations []diff.OperationPair) bool {
	return slices.ContainsFunc(operations, func(pair diff.OperationPair) bool {
		return pair.Base.Security == nil && pair.Revision.Security == nil
	})
}

func globalAnonymousAccessChange(config *Config, access anonymousAccess, revisionSecurity openapi3.SecurityRequirements) SecurityChange {
	if access.base {
		return SecurityChange{
			Id:    APIGlobalSecurityAnonymousAccessRemovedId,
			Level: config.getLogLevel(APIGlobalSecurityAnonymousAccessRemovedId),
			Args:  []any{securityRequirementsString(revisionSecurity)},
		}
	}
	return SecurityChange{
		Id:    APIGlobalSecurityAnonymousAccessAddedId,
		Level: config.getLogLevel(APIGlobalSecurityAnonymousAccessAddedId),
	}
}

func effectiveSecurity(operation *openapi3.Operation, global *openapi3.SecurityRequirements) *openapi3.SecurityRequirements {
	if operation.Security != nil {
		return operation.Security
	}
	return global
}

func operationSecurityChanges(opInfo opInfo, operationsSources *diff.OperationsSourcesMap, base, revision *openapi3.Operation, revisionSecurity *openapi3.SecurityRequirements, securityDiff *diff.SecurityRequirementsDiff, access anonymousAccess) Changes {
	result := make(Changes, 0)

	baseSource := location.SecuritySource(operationsSources, base)
	revisionSource := location.SecuritySource(operationsSources, revision)

	// The change in anonymous access accounts for every alternative and scope
	// that changed with it.
	if access.changed() {
		if access.base {
			return append(result, opInfo.NewApiChange(APISecurityAnonymousAccessRemovedId, []any{securityRequirementsString(*revisionSecurity)}, "").WithSources(baseSource, revisionSource))
		}
		return append(result, opInfo.NewApiChange(APISecurityAnonymousAccessAddedId, nil, "").WithSources(baseSource, revisionSource))
	}

	if securityDiff == nil {
		return result
	}

	// An empty alternative that leaves anonymous access unchanged changes nothing.
	for _, addedSecurity := range securityDiff.Added {
		if len(addedSecurity.Schemes) == 0 {
			continue
		}
		result = append(result, opInfo.NewApiChange(
			APISecurityAddedCheckId,
			[]any{addedSecurity.String()},
			"",
		).WithSources(nil, revisionSource))
	}

	if !access.revisionAcceptsAll() {
		for _, deletedSecurity := range securityDiff.Deleted {
			if len(deletedSecurity.Schemes) == 0 {
				continue
			}
			result = append(result, opInfo.NewApiChange(
				APISecurityRemovedCheckId,
				[]any{deletedSecurity.String()},
				"",
			).WithSources(baseSource, nil))
		}
	}

	for _, updatedSecurity := range securityDiff.Modified {
		for securitySchemeName, updatedSecuritySchemeScopes := range updatedSecurity.Scopes {
			if !access.revisionAcceptsAll() {
				for _, addedScope := range updatedSecuritySchemeScopes.Added {
					result = append(result, opInfo.NewApiChange(
						APISecurityScopeAddedId,
						[]any{addedScope, securitySchemeName},
						"",
					).WithSources(nil, revisionSource))
				}
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

	return result
}
