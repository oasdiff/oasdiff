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
		Args:      []any{"petstore_auth", "http://example.org/api/oauth/dialog", "http://example.new.org/api/oauth/dialog"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth url changed from `http://example.org/api/oauth/dialog` to `http://example.new.org/api/oauth/dialog`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
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
		Args:      []any{"petstore_auth", "", "http://example.new.org/api/oauth/dialog"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth token url changed from `` to `http://example.new.org/api/oauth/dialog`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
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
		Args:      []any{"petstore_auth", "admin:pets"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth scope `admin:pets` was added", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
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
		Args:      []any{"petstore_auth", "admin:pets"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth scope `admin:pets` was removed", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
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
		Args:      []any{"petstore_auth", "read:pets", "read your pets", "grants access to pets (deprecated)"},
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` oauth scope `read:pets` was updated from `read your pets` to `grants access to pets (deprecated)`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
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
		Level:     checker.ERR,
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
