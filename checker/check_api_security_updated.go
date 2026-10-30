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

	APISecurityAnonymousAccessRemovedId = "api-security-anonymous-access-removed"
	APISecurityAnonymousAccessAddedId   = "api-security-anonymous-access-added"

	APISecurityInheritedCommentId          = "api-security-inherited-comment"
	APIGlobalSecurityNotInheritedCommentId = "api-global-security-not-inherited-comment"
)

// checkGlobalSecurity reports changes to the root list when no operation
// inherits it, since there is then no operation to report them on.
func checkGlobalSecurity(diffReport *diff.Diff, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.SecurityDiff == nil {
		return result
	}

	securityContext := diffReport.SecurityContext
	if securityContext != nil && inheritsRootSecurity(securityContext.Operations) {
		return result
	}

	// The document-root "security" field location in each spec; nil when origin
	// tracking is off. Added/scope-added are reported against the revision, the
	// rest against the base, matching the add/remove source convention.
	baseSource := location.SourceFromField(diffReport.SecurityDiff.BaseOrigin, "security")
	revisionSource := location.SourceFromField(diffReport.SecurityDiff.RevisionOrigin, "security")

	// A diff without the security context cannot say whether any operation
	// inherits the list, so the rule's level applies.
	newChange := func(id string, args []any) SecurityChange {
		if securityContext == nil {
			return SecurityChange{Id: id, Level: config.getLogLevel(id), Args: args}
		}
		return SecurityChange{Id: id, Level: INFO, Args: args, Comment: APIGlobalSecurityNotInheritedCommentId}
	}

	// An empty alternative changes only anonymous access, which matters only
	// to operations that inherit the list.
	for _, addedSecurity := range diffReport.SecurityDiff.Added {
		if len(addedSecurity.Schemes) == 0 {
			continue
		}
		result = append(result, newChange(APIGlobalSecurityAddedCheckId, []any{addedSecurity.String()}).WithSources(nil, revisionSource))
	}

	for _, removedSecurity := range diffReport.SecurityDiff.Deleted {
		if len(removedSecurity.Schemes) == 0 {
			continue
		}
		result = append(result, newChange(APIGlobalSecurityRemovedCheckId, []any{removedSecurity.String()}).WithSources(baseSource, nil))
	}

	for _, updatedSecurity := range diffReport.SecurityDiff.Modified {
		for securitySchemeName, updatedSecuritySchemeScopes := range updatedSecurity.Scopes {
			for _, addedScope := range updatedSecuritySchemeScopes.Added {
				result = append(result, newChange(APIGlobalSecurityScopeAddedId, []any{addedScope, securitySchemeName}).WithSources(nil, revisionSource))
			}
			for _, deletedScope := range updatedSecuritySchemeScopes.Deleted {
				result = append(result, newChange(APIGlobalSecurityScopeRemovedId, []any{deletedScope, securitySchemeName}).WithSources(baseSource, nil))
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
			baseSource := location.SecuritySource(operationsSources, operationItem.Base)
			revisionSource := location.SecuritySource(operationsSources, operationItem.Revision)
			result = append(result, operationSecurityChanges(opInfo, baseSource, revisionSource, operationItem.Revision.Security, operationItem.SecurityDiff, declaredAnonymousAccess(operationItem), "")...)
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

		baseSecurity := effectiveSecurity(pair.Base, &securityContext.Base)
		revisionSecurity := effectiveSecurity(pair.Revision, &securityContext.Revision)
		access := effectiveAnonymousAccess(baseSecurity, revisionSecurity)

		comment := ""
		if pair.Base.Security == nil && pair.Revision.Security == nil {
			// The operation's own definition is unchanged.
			comment = APISecurityInheritedCommentId
		}

		handled[diff.Endpoint{Method: pair.Method, Path: pair.Path}] = true
		opInfo := newOpInfo(config, pair.Revision, operationsSources, pair.Method, pair.Path)
		baseSource := inheritedSecuritySource(operationsSources, pair.Base, diffReport.SecurityDiff, true)
		revisionSource := inheritedSecuritySource(operationsSources, pair.Revision, diffReport.SecurityDiff, false)
		securityDiff := diff.GetSecurityRequirementsDiff(baseSecurity, revisionSecurity)
		result = append(result, operationSecurityChanges(opInfo, baseSource, revisionSource, revisionSecurity, securityDiff, access, comment)...)
	}

	return result, handled
}

func inheritsRootSecurity(operations []diff.OperationPair) bool {
	return slices.ContainsFunc(operations, func(pair diff.OperationPair) bool {
		return pair.Base.Security == nil && pair.Revision.Security == nil
	})
}

// inheritedSecuritySource points at the operation's own list, or at the root
// list when the operation inherits it.
func inheritedSecuritySource(operationsSources *diff.OperationsSourcesMap, operation *openapi3.Operation, rootDiff *diff.SecurityRequirementsDiff, base bool) *Source {
	if operation.Security != nil || rootDiff == nil {
		return location.SecuritySource(operationsSources, operation)
	}
	if base {
		return location.SourceFromField(rootDiff.BaseOrigin, "security")
	}
	return location.SourceFromField(rootDiff.RevisionOrigin, "security")
}

func effectiveSecurity(operation *openapi3.Operation, global *openapi3.SecurityRequirements) *openapi3.SecurityRequirements {
	if operation.Security != nil {
		return operation.Security
	}
	return global
}

func operationSecurityChanges(opInfo opInfo, baseSource, revisionSource *Source, revisionSecurity *openapi3.SecurityRequirements, securityDiff *diff.SecurityRequirementsDiff, access anonymousAccess, comment string) Changes {
	result := make(Changes, 0)

	// The change in anonymous access accounts for every alternative and scope
	// that changed with it.
	if access.changed() {
		if access.base {
			return append(result, opInfo.NewApiChange(APISecurityAnonymousAccessRemovedId, []any{securityAlternatives(*revisionSecurity)}, comment).WithSources(baseSource, revisionSource))
		}
		return append(result, opInfo.NewApiChange(APISecurityAnonymousAccessAddedId, nil, comment).WithSources(baseSource, revisionSource))
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
			comment,
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
				comment,
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
						comment,
					).WithSources(nil, revisionSource))
				}
			}
			for _, deletedScope := range updatedSecuritySchemeScopes.Deleted {
				result = append(result, opInfo.NewApiChange(
					APISecurityScopeRemovedId,
					[]any{deletedScope, securitySchemeName},
					comment,
				).WithSources(baseSource, nil))
			}
		}
	}

	return result
}
