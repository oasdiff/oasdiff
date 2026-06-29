package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APIComponentsSecurityRemovedId                  = "api-security-component-removed"
	APIComponentsSecurityAddedId                    = "api-security-component-added"
	APIComponentsSecurityComponentOauthUrlUpdatedId = "api-security-component-oauth-url-changed"
	APIComponentsSecurityTypeUpdatedId              = "api-security-component-type-changed"
	APIComponentsSecurityOauthTokenUrlUpdatedId     = "api-security-component-oauth-token-url-changed"
	APIComponentSecurityOauthScopeAddedId           = "api-security-component-oauth-scope-added"
	APIComponentSecurityOauthScopeRemovedId         = "api-security-component-oauth-scope-removed"
	APIComponentSecurityOauthScopeUpdatedId         = "api-security-component-oauth-scope-changed"
)

const ComponentSecuritySchemes = "securitySchemes"

// checkOAuthUpdates reports oauth flow changes for a modified security scheme.
// baseSource/revisionSource locate the scheme in each spec: added scopes are
// reported against the revision, removed scopes against the base, and in-place
// changes (urls, scope value) against both, matching the source convention used
// elsewhere.
func checkOAuthUpdates(updatedSecurity *diff.SecuritySchemeDiff, updatedSecurityName string, baseSource, revisionSource *Source) Changes {
	result := make(Changes, 0)

	if updatedSecurity.OAuthFlowsDiff == nil {
		return result
	}

	if updatedSecurity.OAuthFlowsDiff.ImplicitDiff == nil {
		return result
	}

	if urlDiff := updatedSecurity.OAuthFlowsDiff.ImplicitDiff.AuthorizationURLDiff; urlDiff != nil {
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityComponentOauthUrlUpdatedId,
			Level:     INFO,
			Args:      []any{updatedSecurityName, urlDiff.From, urlDiff.To},
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, revisionSource))
	}

	if tokenDiff := updatedSecurity.OAuthFlowsDiff.ImplicitDiff.TokenURLDiff; tokenDiff != nil {
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityOauthTokenUrlUpdatedId,
			Level:     INFO,
			Args:      []any{updatedSecurityName, tokenDiff.From, tokenDiff.To},
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, revisionSource))
	}

	if scopesDiff := updatedSecurity.OAuthFlowsDiff.ImplicitDiff.ScopesDiff; scopesDiff != nil {
		for _, addedScope := range scopesDiff.Added {
			result = append(result, ComponentChange{
				Id:        APIComponentSecurityOauthScopeAddedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, addedScope},
				Component: ComponentSecuritySchemes,
			}.WithSources(nil, revisionSource))
		}

		for _, removedScope := range scopesDiff.Deleted {
			result = append(result, ComponentChange{
				Id:        APIComponentSecurityOauthScopeRemovedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, removedScope},
				Component: ComponentSecuritySchemes,
			}.WithSources(baseSource, nil))
		}

		for name, modifiedScope := range scopesDiff.Modified {
			result = append(result, ComponentChange{
				Id:        APIComponentSecurityOauthScopeUpdatedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, name, modifiedScope.From, modifiedScope.To},
				Component: ComponentSecuritySchemes,
			}.WithSources(baseSource, revisionSource))
		}

	}

	return result
}

func APIComponentsSecurityUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	if diffReport.ComponentsDiff == nil {
		return result
	}

	if diffReport.ComponentsDiff.SecuritySchemesDiff == nil {
		return result
	}

	for _, updatedSecurity := range diffReport.ComponentsDiff.SecuritySchemesDiff.Added {
		var revisionSource *Source
		if ref := diffReport.ComponentsDiff.SecuritySchemesDiff.Revision[updatedSecurity]; ref != nil && ref.Value != nil {
			revisionSource = sourceFromOrigin(ref.Value.Origin)
		}
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityAddedId,
			Level:     INFO,
			Args:      []any{updatedSecurity},
			Component: ComponentSecuritySchemes,
		}.WithSources(nil, revisionSource))
	}

	for _, updatedSecurity := range diffReport.ComponentsDiff.SecuritySchemesDiff.Deleted {
		var baseSource *Source
		if ref := diffReport.ComponentsDiff.SecuritySchemesDiff.Base[updatedSecurity]; ref != nil && ref.Value != nil {
			baseSource = sourceFromOrigin(ref.Value.Origin)
		}
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityRemovedId,
			Level:     INFO,
			Args:      []any{updatedSecurity},
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, nil))
	}

	for updatedSecurityName, updatedSecurity := range diffReport.ComponentsDiff.SecuritySchemesDiff.Modified {
		var baseSource, revisionSource *Source
		if ref := diffReport.ComponentsDiff.SecuritySchemesDiff.Base[updatedSecurityName]; ref != nil && ref.Value != nil {
			baseSource = sourceFromOrigin(ref.Value.Origin)
		}
		if ref := diffReport.ComponentsDiff.SecuritySchemesDiff.Revision[updatedSecurityName]; ref != nil && ref.Value != nil {
			revisionSource = sourceFromOrigin(ref.Value.Origin)
		}

		result = append(result, checkOAuthUpdates(updatedSecurity, updatedSecurityName, baseSource, revisionSource)...)

		if updatedSecurity.TypeDiff != nil {
			result = append(result, ComponentChange{
				Id:        APIComponentsSecurityTypeUpdatedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, updatedSecurity.TypeDiff.From, updatedSecurity.TypeDiff.To},
				Component: ComponentSecuritySchemes,
			}.WithSources(baseSource, revisionSource))
		}
	}

	return result
}
