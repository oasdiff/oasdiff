package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: changing security component oauth's url
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

// CL: changing security component token url
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

// CL: changing security component type
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
		Level:     checker.INFO,
		Component: checker.ComponentSecuritySchemes,
	}, errs[0])
	require.Equal(t, "the component security scheme `petstore_auth` type changed from `oauth2` to `http`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: adding a new security component
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

// CL: removing a new security component
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

// CL: adding a new oauth security scope
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

// CL: removing a new oauth security scope
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

// CL: removing a new oauth security scope
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
