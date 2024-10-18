package checker

import (
	"github.com/tufin/oasdiff/diff"
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

func checkOAuthUpdates(updatedSecurity *diff.SecuritySchemeDiff, updatedSecurityName string) Changes {
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
		})
	}

	if tokenDiff := updatedSecurity.OAuthFlowsDiff.ImplicitDiff.TokenURLDiff; tokenDiff != nil {
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityOauthTokenUrlUpdatedId,
			Level:     INFO,
			Args:      []any{updatedSecurityName, tokenDiff.From, tokenDiff.To},
			Component: ComponentSecuritySchemes,
		})
	}

	if scopesDiff := updatedSecurity.OAuthFlowsDiff.ImplicitDiff.ScopesDiff; scopesDiff != nil {
		for _, addedScope := range scopesDiff.Added {
			result = append(result, ComponentChange{
				Id:        APIComponentSecurityOauthScopeAddedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, addedScope},
				Component: ComponentSecuritySchemes,
			})
		}

		for _, removedScope := range scopesDiff.Deleted {
			result = append(result, ComponentChange{
				Id:        APIComponentSecurityOauthScopeRemovedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, removedScope},
				Component: ComponentSecuritySchemes,
			})
		}

		for name, modifiedScope := range scopesDiff.Modified {
			result = append(result, ComponentChange{
				Id:        APIComponentSecurityOauthScopeUpdatedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, name, modifiedScope.From, modifiedScope.To},
				Component: ComponentSecuritySchemes,
			})
		}

	}

	return result
}

func APIComponentsSecurityUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.ComponentsDiff.SecuritySchemesDiff == nil {
		return result
	}

	for _, updatedSecurity := range diffReport.ComponentsDiff.SecuritySchemesDiff.Added {
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityAddedId,
			Level:     INFO,
			Args:      []any{updatedSecurity},
			Component: ComponentSecuritySchemes,
		})
	}

	for _, updatedSecurity := range diffReport.ComponentsDiff.SecuritySchemesDiff.Deleted {
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityRemovedId,
			Level:     INFO,
			Args:      []any{updatedSecurity},
			Component: ComponentSecuritySchemes,
		})
	}

	for updatedSecurityName, updatedSecurity := range diffReport.ComponentsDiff.SecuritySchemesDiff.Modified {
		result = append(result, checkOAuthUpdates(updatedSecurity, updatedSecurityName)...)

		if updatedSecurity.TypeDiff != nil {
			result = append(result, ComponentChange{
				Id:        APIComponentsSecurityTypeUpdatedId,
				Level:     INFO,
				Args:      []any{updatedSecurityName, updatedSecurity.TypeDiff.From, updatedSecurity.TypeDiff.To},
				Component: ComponentSecuritySchemes,
			})
		}
	}

	return result
}
