package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// changing security component oauth's url
func TestComponentSecurityOauthURLUpdated(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Implicit.AuthorizationURL = "http://example.new.org/api/oauth/dialog"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityComponentOauthUrlUpdatedId,
		Args:      []any{"petstore_auth", "implicit", "http://example.org/api/oauth/dialog", "http://example.new.org/api/oauth/dialog"},
		Level:     checker.WARN,
		Comment:   checker.SecuritySchemeURLChangedCommentId,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth `implicit` flow authorization url changed from `http://example.org/api/oauth/dialog` to `http://example.new.org/api/oauth/dialog`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing security component token url
func TestComponentSecurityOauthTokenUpdated(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Implicit.TokenURL = "http://example.new.org/api/oauth/dialog"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityOauthTokenUrlUpdatedId,
		Args:      []any{"petstore_auth", "implicit", "", "http://example.new.org/api/oauth/dialog"},
		Level:     checker.WARN,
		Comment:   checker.SecuritySchemeURLChangedCommentId,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth `implicit` flow token url changed from `` to `http://example.new.org/api/oauth/dialog`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing security component type
func TestComponentSecurityTypeUpdated(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Type = "http"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityTypeUpdatedId,
		Args:      []any{"petstore_auth", "oauth2", "http"},
		Level:     checker.ERR,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` type changed from `oauth2` to `http`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// adding a new security component
func TestComponentSecurityAdded(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityAddedId,
		Args:      []any{"BasicAuth"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `BasicAuth` was added", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing a new security component
func TestComponentSecurityRemoved(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityRemovedId,
		Args:      []any{"BasicAuth"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `BasicAuth` was removed", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// adding a new oauth security scope
func TestComponentSecurityOauthScopeAdded(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Implicit.Scopes["admin:pets"] = "grants access to admin operations"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentSecurityOauthScopeAddedId,
		Args:      []any{"petstore_auth", "implicit", "admin:pets"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth `implicit` flow scope `admin:pets` was added", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// adding an oauth scope, with source tracking: the change points at the
// revision scheme's location.
func TestComponentSecurityOauthScopeAdded_WithSources(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)

	s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Implicit.Scopes["admin:pets"] = "grants access to admin operations"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.APIComponentSecurityOauthScopeAddedId)

	require.NotEmpty(t, errs[0].GetRevisionSource())
	require.Equal(t, "../data/checker/component_security_updated_base.yaml", errs[0].GetRevisionSource().File)
	require.Equal(t, 29, errs[0].GetRevisionSource().Line)
	require.Empty(t, errs[0].GetBaseSource())
}

// removing a new oauth security scope
func TestComponentSecurityOauthScopeRemoved(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)

	// Add to s1 so that it's deletion is identified
	s1.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Implicit.Scopes["admin:pets"] = "grants access to admin operations"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        "api-security-component-oauth-scope-removed",
		Args:      []any{"petstore_auth", "implicit", "admin:pets"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth `implicit` flow scope `admin:pets` was removed", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing a new oauth security scope
func TestComponentSecurityOauthScopeUpdated(t *testing.T) {
	s1, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_updated_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Implicit.Scopes["read:pets"] = "grants access to pets (deprecated)"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        "api-security-component-oauth-scope-changed",
		Args:      []any{"petstore_auth", "implicit", "read:pets", "read your pets", "grants access to pets (deprecated)"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth `implicit` flow scope `read:pets` was updated from `read your pets` to `grants access to pets (deprecated)`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func checkWireFields(t *testing.T, modify func(schemes openapi3.SecuritySchemes)) checker.Changes {
	t.Helper()
	s1, err := open("../data/checker/component_security_wire_fields_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/component_security_wire_fields_base.yaml")
	require.NoError(t, err)

	modify(s2.Spec.Components.SecuritySchemes)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
}

// changing the apiKey name moves the key to a different header
func TestComponentSecurityApiKeyNameUpdated(t *testing.T) {
	errs := checkWireFields(t, func(schemes openapi3.SecuritySchemes) {
		schemes["api_key"].Value.Name = "X-Token"
	})
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityApiKeyNameUpdatedId,
		Args:      []any{"api_key", "X-API-Key", "X-Token"},
		Level:     checker.ERR,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `api_key` api key name changed from `X-API-Key` to `X-Token`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing the apiKey location moves the key, e.g. from a header to the query
func TestComponentSecurityApiKeyInUpdated(t *testing.T) {
	errs := checkWireFields(t, func(schemes openapi3.SecuritySchemes) {
		schemes["api_key"].Value.In = "query"
	})
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityApiKeyInUpdatedId,
		Args:      []any{"api_key", "header", "query"},
		Level:     checker.ERR,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `api_key` api key location changed from `header` to `query`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing the http scheme changes the Authorization header format
func TestComponentSecurityHttpSchemeUpdated(t *testing.T) {
	errs := checkWireFields(t, func(schemes openapi3.SecuritySchemes) {
		schemes["basic_auth"].Value.Scheme = "bearer"
	})
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityHttpSchemeUpdatedId,
		Args:      []any{"basic_auth", "basic", "bearer"},
		Level:     checker.ERR,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `basic_auth` http scheme changed from `basic` to `bearer`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// HTTP authentication scheme names are case-insensitive (RFC 9110, section 11.1)
func TestComponentSecurityHttpSchemeCaseChanged(t *testing.T) {
	errs := checkWireFields(t, func(schemes openapi3.SecuritySchemes) {
		schemes["basic_auth"].Value.Scheme = "Basic"
	})
	require.Empty(t, errs)
}

func TestComponentSecurityBearerFormatUpdated(t *testing.T) {
	errs := checkWireFields(t, func(schemes openapi3.SecuritySchemes) {
		schemes["bearer_auth"].Value.BearerFormat = "opaque"
	})
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityBearerFormatUpdatedId,
		Args:      []any{"bearer_auth", "JWT", "opaque"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `bearer_auth` bearer format changed from `JWT` to `opaque`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func TestComponentSecurityOpenIdConnectUrlUpdated(t *testing.T) {
	errs := checkWireFields(t, func(schemes openapi3.SecuritySchemes) {
		schemes["oidc"].Value.OpenIdConnectUrl = "https://auth.example.com/.well-known/openid-configuration"
	})
	require.Len(t, errs, 1)
	require.Equal(t, checker.ComponentChange{
		Id:        checker.APIComponentsSecurityOpenIdConnectUrlUpdatedId,
		Args:      []any{"oidc", "https://example.com/.well-known/openid-configuration", "https://auth.example.com/.well-known/openid-configuration"},
		Level:     checker.WARN,
		Comment:   checker.SecuritySchemeURLChangedCommentId,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `oidc` OpenID Connect url changed from `https://example.com/.well-known/openid-configuration` to `https://auth.example.com/.well-known/openid-configuration`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// a type change replaces the whole authentication method, so the fields that
// come and go with it are reported as the type change alone
func TestComponentSecurityTypeUpdated_WireFieldsNotRepeated(t *testing.T) {
	errs := checkWireFields(t, func(schemes openapi3.SecuritySchemes) {
		schemes["api_key"].Value = &openapi3.SecurityScheme{Type: "http", Scheme: "bearer"}
	})
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIComponentsSecurityTypeUpdatedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// URLs that name the same resource under RFC 3986 normalization are not a
// change; anything else is
func TestComponentSecurityURLEquivalence(t *testing.T) {
	tests := []struct {
		from, to string
		same     bool
	}{
		{"https://example.com/.well-known/openid-configuration", "HTTPS://Example.COM/.well-known/openid-configuration", true},
		{"https://example.com/.well-known/openid-configuration", "https://example.com:443/.well-known/openid-configuration", true},
		{"http://example.com/token", "http://example.com:80/token", true},
		{"https://example.com", "https://example.com/", true},
		{"https://example.com/a%2fb", "https://example.com/a%2Fb", true},
		{"https://example.com/~user", "https://example.com/%7Euser", true},
		{"https://[::1]:443/token", "https://[::1]/token", true},
		{"/token", "/token", true},
		{"https://example.com/.well-known/openid-configuration", "https://example.com/.well-known/OpenID-configuration", false},
		{"https://example.com/token", "http://example.com/token", false},
		{"https://example.com/token", "https://example.com:8443/token", false},
		{"https://example.com/token", "https://example.com/token/", false},
		{"https://example.com/token", "https://example.com/token?tenant=a", false},
		{"https://example.com/a/b", "https://example.com/a%2Fb", false},
	}

	for _, tc := range tests {
		t.Run(tc.from+" to "+tc.to, func(t *testing.T) {
			s1, err := open("../data/checker/component_security_wire_fields_base.yaml")
			require.NoError(t, err)
			s2, err := open("../data/checker/component_security_wire_fields_base.yaml")
			require.NoError(t, err)
			s1.Spec.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl = tc.from
			s2.Spec.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl = tc.to

			d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
			require.NoError(t, err)
			errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)
			if tc.same {
				require.Empty(t, errs)
			} else {
				requireSingleChange(t, errs, checker.APIComponentsSecurityOpenIdConnectUrlUpdatedId)
			}
		})
	}
}

// every OAuth flow is checked, and each finding names its flow
func TestComponentSecurityOauthFlows(t *testing.T) {
	flow := func() *openapi3.OAuthFlow {
		return &openapi3.OAuthFlow{
			AuthorizationURL: "https://example.com/authorize",
			TokenURL:         "https://example.com/token",
			RefreshURL:       "https://example.com/refresh",
			Scopes:           map[string]string{"read": "read pets"},
		}
	}
	setFlow := func(flows *openapi3.OAuthFlows, name string, f *openapi3.OAuthFlow) {
		switch name {
		case "implicit":
			flows.Implicit = f
		case "password":
			flows.Password = f
		case "clientCredentials":
			flows.ClientCredentials = f
		case "authorizationCode":
			flows.AuthorizationCode = f
		}
	}

	for _, name := range []string{"implicit", "password", "clientCredentials", "authorizationCode"} {
		t.Run(name, func(t *testing.T) {
			s1, err := open("../data/checker/component_security_updated_base.yaml")
			require.NoError(t, err)
			s2, err := open("../data/checker/component_security_updated_base.yaml")
			require.NoError(t, err)
			s1.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows = &openapi3.OAuthFlows{}
			s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows = &openapi3.OAuthFlows{}

			setFlow(s1.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows, name, flow())
			revision := flow()
			revision.AuthorizationURL = "https://auth.example.com/authorize"
			revision.TokenURL = "https://auth.example.com/token"
			revision.RefreshURL = "https://auth.example.com/refresh"
			revision.Scopes["write"] = "write pets"
			setFlow(s2.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows, name, revision)

			d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
			require.NoError(t, err)
			errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIComponentsSecurityUpdatedCheck), d, osm, checker.INFO)

			found := map[string]checker.Level{}
			for _, change := range errs {
				require.Equal(t, name, change.GetArgs()[1], change.GetId())
				found[change.GetId()] = change.GetLevel()
			}
			require.Equal(t, map[string]checker.Level{
				checker.APIComponentsSecurityComponentOauthUrlUpdatedId: checker.WARN,
				checker.APIComponentsSecurityOauthTokenUrlUpdatedId:     checker.WARN,
				checker.APIComponentsSecurityOauthRefreshUrlUpdatedId:   checker.WARN,
				checker.APIComponentSecurityOauthScopeAddedId:           checker.INFO,
			}, found)
		})
	}
}

func TestComponentSecurityOauthRefreshUrlText(t *testing.T) {
	require.Equal(t,
		"the component security scheme `oauth` oauth `password` flow refresh url changed from `https://a.example.com` to `https://b.example.com`",
		checker.ComponentChange{
			Id:   checker.APIComponentsSecurityOauthRefreshUrlUpdatedId,
			Args: []any{"oauth", "password", "https://a.example.com", "https://b.example.com"},
		}.GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t,
		"This is a warning because the spec does not say whether credentials obtained through the previous URL are still accepted.",
		checker.NewDefaultLocalizer()(checker.SecuritySchemeURLChangedCommentId))
}
