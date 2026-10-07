package checker

import (
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

func checkGlobalSecurity(diffReport *diff.Diff) Changes {
	result := make(Changes, 0)
	if diffReport.SecurityDiff == nil {
		return result
	}

	// The document-root "security" field location in each spec; nil when origin
	// tracking is off. Added/scope-added are reported against the revision, the
	// rest against the base, matching the add/remove source convention.
	baseSource := location.SourceFromField(diffReport.SecurityDiff.BaseOrigin, "security")
	revisionSource := location.SourceFromField(diffReport.SecurityDiff.RevisionOrigin, "security")

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

			if operationItem.SecurityDiff == nil || handled[operationItem.Revision] {
				continue
			}

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			access := anonymousAccess{
				// An operation that inherits the root list on either side was
				// handled above when the diff carries that list; otherwise its
				// effective security is unknown.
				known:    operationItem.Base.Security != nil && operationItem.Revision.Security != nil,
				base:     allowsAnonymous(operationItem.Base.Security),
				revision: allowsAnonymous(operationItem.Revision.Security),
			}
			result = append(result, operationSecurityChanges(opInfo, operationsSources, operationItem.Base, operationItem.Revision, operationItem.SecurityDiff, access)...)
		}
	}

	return result
}

// checkInheritedSecurity reports the operations whose security comes, on at
// least one side, from the document-root list. It returns the revision
// operations it reported on, so the caller skips their own security diff,
// which does not account for the inherited list.
func checkInheritedSecurity(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) (Changes, map[*openapi3.Operation]bool) {
	result := make(Changes, 0)
	handled := map[*openapi3.Operation]bool{}

	securityContext := diffReport.SecurityContext
	if securityContext == nil {
		return result, handled
	}

	globalAccessChanged := false
	for _, pair := range securityContext.Operations {
		if pair.Base.Security != nil && pair.Revision.Security != nil {
			continue
		}
		if !inStabilityScope(config, pair.Revision) {
			continue
		}

		baseSecurity := effectiveSecurity(pair.Base, &securityContext.Base)
		revisionSecurity := effectiveSecurity(pair.Revision, &securityContext.Revision)
		access := anonymousAccess{
			known:    true,
			base:     allowsAnonymous(baseSecurity),
			revision: allowsAnonymous(revisionSecurity),
		}

		if pair.Base.Security == nil && pair.Revision.Security == nil {
			// Only the root list changed for this operation, so the change is
			// reported once, against the root list, rather than per operation.
			globalAccessChanged = globalAccessChanged || access.changed()
			continue
		}

		handled[pair.Revision] = true
		opInfo := newOpInfo(config, pair.Revision, operationsSources, pair.Method, pair.Path)
		securityDiff := diff.GetSecurityRequirementsDiff(baseSecurity, revisionSecurity)
		result = append(result, operationSecurityChanges(opInfo, operationsSources, pair.Base, pair.Revision, securityDiff, access)...)
	}

	if globalAccessChanged {
		id := APIGlobalSecurityAnonymousAccessAddedId
		if allowsAnonymous(&securityContext.Base) {
			id = APIGlobalSecurityAnonymousAccessRemovedId
		}
		var baseSource, revisionSource *Source
		if diffReport.SecurityDiff != nil {
			baseSource = location.SourceFromField(diffReport.SecurityDiff.BaseOrigin, "security")
			revisionSource = location.SourceFromField(diffReport.SecurityDiff.RevisionOrigin, "security")
		}
		result = append(result, SecurityChange{
			Id:    id,
			Level: config.getLogLevel(id),
		}.WithSources(baseSource, revisionSource))
	}

	return result, handled
}

// anonymousAccess records whether an operation accepts unauthenticated
// requests before and after the change.
type anonymousAccess struct {
	known    bool
	base     bool
	revision bool
}

func (access anonymousAccess) changed() bool {
	return access.known && access.base != access.revision
}

// Once the revision accepts unauthenticated requests it accepts every request,
// so nothing removed from its requirements can reject a client.
func (access anonymousAccess) revisionAcceptsAll() bool {
	return access.known && access.revision
}

func effectiveSecurity(operation *openapi3.Operation, global *openapi3.SecurityRequirements) *openapi3.SecurityRequirements {
	if operation.Security != nil {
		return operation.Security
	}
	return global
}

// allowsAnonymous reports whether a security list accepts a request without
// credentials: an empty list requires nothing, and an empty alternative (`{}`)
// is satisfied by any request.
func allowsAnonymous(securityRequirements *openapi3.SecurityRequirements) bool {
	if securityRequirements == nil || len(*securityRequirements) == 0 {
		return true
	}
	for _, securityRequirement := range *securityRequirements {
		if len(securityRequirement) == 0 {
			return true
		}
	}
	return false
}

func inStabilityScope(config *Config, operation *openapi3.Operation) bool {
	if config == nil {
		return true
	}
	stability, err := getStabilityLevel(operation.Extensions)
	if err != nil {
		return true
	}
	return config.StabilityLevel.IsIncluded(stability)
}

func operationSecurityChanges(opInfo opInfo, operationsSources *diff.OperationsSourcesMap, base, revision *openapi3.Operation, securityDiff *diff.SecurityRequirementsDiff, access anonymousAccess) Changes {
	result := make(Changes, 0)

	baseSource := location.SecuritySource(operationsSources, base)
	revisionSource := location.SecuritySource(operationsSources, revision)

	if access.changed() {
		id := APISecurityAnonymousAccessAddedId
		if access.base {
			id = APISecurityAnonymousAccessRemovedId
		}
		result = append(result, opInfo.NewApiChange(id, nil, "").WithSources(baseSource, revisionSource))
	}

	if securityDiff == nil {
		return result
	}

	// An empty alternative is reported as a change in anonymous access above.
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
