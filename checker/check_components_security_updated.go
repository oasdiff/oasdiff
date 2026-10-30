package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/checker/location"
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
	APIComponentsSecurityApiKeyNameUpdatedId        = "api-security-component-api-key-name-changed"
	APIComponentsSecurityApiKeyInUpdatedId          = "api-security-component-api-key-in-changed"
	APIComponentsSecurityHttpSchemeUpdatedId        = "api-security-component-http-scheme-changed"
	APIComponentsSecurityBearerFormatUpdatedId      = "api-security-component-bearer-format-changed"
	APIComponentsSecurityOpenIdConnectUrlUpdatedId  = "api-security-component-openid-connect-url-changed"
	APIComponentsSecurityOauthRefreshUrlUpdatedId   = "api-security-component-oauth-refresh-url-changed"

	SecuritySchemeURLChangedCommentId = "security-scheme-url-changed-warn-comment"
)

const ComponentSecuritySchemes = "securitySchemes"

// checkOAuthUpdates reports oauth flow changes for a modified security scheme.
// baseSource/revisionSource locate the scheme in each spec: added scopes are
// reported against the revision, removed scopes against the base, and in-place
// changes (urls, scope value) against both, matching the source convention used
// elsewhere.
func checkOAuthUpdates(updatedSecurity *diff.SecuritySchemeDiff, updatedSecurityName string, baseSource, revisionSource *Source) Changes {
	result := make(Changes, 0)

	flowsDiff := updatedSecurity.OAuthFlowsDiff
	if flowsDiff == nil {
		return result
	}

	flows := []struct {
		name string
		diff *diff.OAuthFlowDiff
	}{
		{"implicit", flowsDiff.ImplicitDiff},
		{"password", flowsDiff.PasswordDiff},
		{"clientCredentials", flowsDiff.ClientCredentialsDiff},
		{"authorizationCode", flowsDiff.AuthorizationCodeDiff},
	}
	for _, flow := range flows {
		if flow.diff != nil {
			result = append(result, checkOAuthFlowUpdates(flow.diff, updatedSecurityName, flow.name, baseSource, revisionSource)...)
		}
	}

	return result
}

func checkOAuthFlowUpdates(flowDiff *diff.OAuthFlowDiff, updatedSecurityName, flowName string, baseSource, revisionSource *Source) Changes {
	result := make(Changes, 0)

	urlChange := func(id string, urlDiff *diff.ValueDiff) {
		if urlDiff == nil || sameURL(urlDiff.From, urlDiff.To) {
			return
		}
		result = append(result, ComponentChange{
			Id:        id,
			Level:     WARN,
			Args:      []any{updatedSecurityName, flowName, urlDiff.From, urlDiff.To},
			Comment:   SecuritySchemeURLChangedCommentId,
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, revisionSource))
	}
	urlChange(APIComponentsSecurityComponentOauthUrlUpdatedId, flowDiff.AuthorizationURLDiff)
	urlChange(APIComponentsSecurityOauthTokenUrlUpdatedId, flowDiff.TokenURLDiff)
	urlChange(APIComponentsSecurityOauthRefreshUrlUpdatedId, flowDiff.RefreshURLDiff)

	scopesDiff := flowDiff.ScopesDiff
	if scopesDiff == nil {
		return result
	}

	for _, addedScope := range scopesDiff.Added {
		result = append(result, ComponentChange{
			Id:        APIComponentSecurityOauthScopeAddedId,
			Level:     INFO,
			Args:      []any{updatedSecurityName, flowName, addedScope},
			Component: ComponentSecuritySchemes,
		}.WithSources(nil, revisionSource))
	}

	for _, removedScope := range scopesDiff.Deleted {
		result = append(result, ComponentChange{
			Id:        APIComponentSecurityOauthScopeRemovedId,
			Level:     INFO,
			Args:      []any{updatedSecurityName, flowName, removedScope},
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, nil))
	}

	for name, modifiedScope := range scopesDiff.Modified {
		result = append(result, ComponentChange{
			Id:        APIComponentSecurityOauthScopeUpdatedId,
			Level:     INFO,
			Args:      []any{updatedSecurityName, flowName, name, modifiedScope.From, modifiedScope.To},
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, revisionSource))
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
			revisionSource = location.SourceFromOrigin(ref.Value.Origin)
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
			baseSource = location.SourceFromOrigin(ref.Value.Origin)
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
			baseSource = location.SourceFromOrigin(ref.Value.Origin)
		}
		if ref := diffReport.ComponentsDiff.SecuritySchemesDiff.Revision[updatedSecurityName]; ref != nil && ref.Value != nil {
			revisionSource = location.SourceFromOrigin(ref.Value.Origin)
		}

		result = append(result, checkOAuthUpdates(updatedSecurity, updatedSecurityName, baseSource, revisionSource)...)

		if updatedSecurity.TypeDiff != nil {
			result = append(result, ComponentChange{
				Id:        APIComponentsSecurityTypeUpdatedId,
				Level:     ERR,
				Args:      []any{updatedSecurityName, updatedSecurity.TypeDiff.From, updatedSecurity.TypeDiff.To},
				Component: ComponentSecuritySchemes,
			}.WithSources(baseSource, revisionSource))
		} else {
			// under a type change, the old type's fields disappear and the new
			// type's appear; reporting each would repeat the type change
			result = append(result, checkWireFieldUpdates(updatedSecurity, updatedSecurityName, baseSource, revisionSource)...)
		}
	}

	return result
}

// checkWireFieldUpdates reports changes to the fields that tell a client how
// to send its credentials.
func checkWireFieldUpdates(updatedSecurity *diff.SecuritySchemeDiff, updatedSecurityName string, baseSource, revisionSource *Source) Changes {
	result := make(Changes, 0)

	report := func(id string, level Level, valueDiff *diff.ValueDiff) {
		if valueDiff == nil {
			return
		}
		result = append(result, ComponentChange{
			Id:        id,
			Level:     level,
			Args:      []any{updatedSecurityName, valueDiff.From, valueDiff.To},
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, revisionSource))
	}

	report(APIComponentsSecurityApiKeyNameUpdatedId, ERR, updatedSecurity.NameDiff)
	report(APIComponentsSecurityApiKeyInUpdatedId, ERR, updatedSecurity.InDiff)
	// HTTP authentication scheme names are case-insensitive (RFC 9110, section 11.1)
	if schemeDiff := updatedSecurity.SchemeDiff; schemeDiff != nil && !strings.EqualFold(interfaceToString(schemeDiff.From), interfaceToString(schemeDiff.To)) {
		report(APIComponentsSecurityHttpSchemeUpdatedId, ERR, schemeDiff)
	}
	report(APIComponentsSecurityBearerFormatUpdatedId, INFO, updatedSecurity.BearerFormatDiff)
	if urlDiff := updatedSecurity.OpenIDConnectURLDiff; urlDiff != nil && !sameURL(urlDiff.From, urlDiff.To) {
		result = append(result, ComponentChange{
			Id:        APIComponentsSecurityOpenIdConnectUrlUpdatedId,
			Level:     WARN,
			Args:      []any{updatedSecurityName, urlDiff.From, urlDiff.To},
			Comment:   SecuritySchemeURLChangedCommentId,
			Component: ComponentSecuritySchemes,
		}.WithSources(baseSource, revisionSource))
	}

	return result
}
