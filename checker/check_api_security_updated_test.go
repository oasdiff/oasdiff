package checker_test

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// adding a new global security to the API
func TestAPIGlobalSecurityyAdded(t *testing.T) {
	s1, err := open("../data/checker/api_security_global_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	requireChange(t, errs, checker.APIGlobalSecurityAnonymousAccessRemovedId)
	added := requireChange(t, errs, checker.APIGlobalSecurityAddedCheckId)
	require.Equal(t, checker.SecurityChange{
		Id:    checker.APIGlobalSecurityAddedCheckId,
		Args:  []any{"petstore_auth: [read:pets, write:pets]"},
		Level: checker.INFO,
	}, added)
	require.Equal(t, "the security scheme `petstore_auth: [read:pets, write:pets]` was added to the API", added.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing the only global security makes the API anonymous, which is
// reported as that instead of as the removed requirement
func TestAPIGlobalSecurityyDeleted(t *testing.T) {
	s1, err := open("../data/checker/api_security_global_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.APIGlobalSecurityAnonymousAccessAddedId)
}

// removing a security scope from an API global security
func TestAPIGlobalSecurityScopeRemoved(t *testing.T) {
	s1, err := open("../data/checker/api_security_global_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_added_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Security[0]["petstore_auth"] = s2.Spec.Security[0]["petstore_auth"][:1]
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.SecurityChange{
		Id:    checker.APIGlobalSecurityScopeRemovedId,
		Args:  []any{"read:pets", "petstore_auth"},
		Level: checker.INFO,
	}, errs[0])
	require.Equal(t, "the security scope `read:pets` was removed from the global security scheme `petstore_auth`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// adding a security scope from an API global security
func TestAPIGlobalSecurityScopeAdded(t *testing.T) {
	s1, err := open("../data/checker/api_security_global_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_added_revision.yaml")
	require.NoError(t, err)

	s1.Spec.Security[0]["petstore_auth"] = s2.Spec.Security[0]["petstore_auth"][:1]
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.SecurityChange{
		Id:    checker.APIGlobalSecurityScopeAddedId,
		Args:  []any{"read:pets", "petstore_auth"},
		Level: checker.ERR,
	}, errs[0])
	require.Equal(t, "the security scope `read:pets` was added to the global security scheme `petstore_auth`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// replacing the top-level security requirement rejects clients of an operation that inherits it
func TestAPIGlobalSecurityReplaced(t *testing.T) {
	changes := securityChanges(t, "api_security_global_api_key", "api_security_global_bearer")
	require.Len(t, changes, 2)
	require.Equal(t, checker.ERR, requireChange(t, changes, checker.APIGlobalSecurityRemovedCheckId).GetLevel())
	require.Equal(t, checker.INFO, requireChange(t, changes, checker.APIGlobalSecurityAddedCheckId).GetLevel())
}

// an operation with its own security does not use the top-level list
func TestAPIGlobalSecurityReplacedNotInherited(t *testing.T) {
	changes := securityChanges(t, "api_security_global_api_key_op_api_key", "api_security_global_bearer_op_api_key")
	require.Len(t, changes, 2)
	removed := requireChange(t, changes, checker.APIGlobalSecurityRemovedCheckId)
	require.Equal(t, checker.INFO, removed.GetLevel())
	require.Equal(t, "This is not breaking because no checked operation inherits the top-level security list; each declares its own security", removed.GetComment(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.INFO, requireChange(t, changes, checker.APIGlobalSecurityAddedCheckId).GetLevel())
}

// adding a new security to the API endpoint
func TestAPISecurityAdded(t *testing.T) {
	s1, err := open("../data/checker/api_security_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	requireChange(t, errs, checker.APISecurityAnonymousAccessRemovedId)
	added := requireChange(t, errs, checker.APISecurityAddedCheckId)
	requireApiChange(t, checker.ApiChange{
		Id:        checker.APISecurityAddedCheckId,
		Args:      []any{"petstore_auth: [read:pets, write:pets]"},
		Operation: "POST",
		Path:      "/subscribe",
		Source:    load.NewSource("../data/checker/api_security_added_revision.yaml"),
	}, added)
	require.Equal(t, "the endpoint scheme security `petstore_auth: [read:pets, write:pets]` was added to the API", added.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing the only security requirement of an API endpoint makes it anonymous,
// which accepts every request, so the removed requirement is not an error
func TestAPISecurityDeleted(t *testing.T) {
	s1, err := open("../data/checker/api_security_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:        checker.APISecurityAnonymousAccessAddedId,
		Operation: "POST",
		Path:      "/subscribe",
		Source:    load.NewSource("../data/checker/api_security_added_base.yaml"),
	}, errs)
	require.Equal(t, "the endpoint now allows anonymous access", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing a security scope from an API endpoint security
func TestAPISecurityScopeRemoved(t *testing.T) {
	s1, err := open("../data/checker/api_security_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:        checker.APISecurityScopeRemovedId,
		Args:      []any{"read:pets", "petstore_auth"},
		Operation: "POST",
		Path:      "/subscribe",
		Source:    load.NewSource("../data/checker/api_security_updated_revision.yaml"),
	}, errs)
	require.Equal(t, "the security scope `read:pets` was removed from the endpoint's security scheme `petstore_auth`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// adding a security scope to an API endpoint security
func TestAPISecurityScopeAdded(t *testing.T) {
	s1, err := open("../data/checker/api_security_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:        checker.APISecurityScopeAddedId,
		Args:      []any{"read:pets", "petstore_auth"},
		Operation: "POST",
		Path:      "/subscribe",
		Source:    load.NewSource("../data/checker/api_security_updated_base.yaml"),
	}, errs)
	require.Equal(t, "the security scope `read:pets` was added to the endpoint's security scheme `petstore_auth`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing a global security, with source tracking: the change points at the
// document-root "security" field of the base spec.
func TestAPIGlobalSecurityDeleted_WithSources(t *testing.T) {
	s1, err := open("../data/checker/api_security_global_api_key.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_bearer.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	removed := requireChange(t, errs, checker.APIGlobalSecurityRemovedCheckId)

	require.NotEmpty(t, removed.GetBaseSource())
	require.Equal(t, "../data/checker/api_security_global_api_key.yaml", removed.GetBaseSource().File)
	require.Equal(t, 5, removed.GetBaseSource().Line)
	require.Empty(t, removed.GetRevisionSource())
}

// making the API anonymous, with source tracking: the change points at the
// document-root "security" field of the base spec.
func TestAPIGlobalSecurityAnonymousAccessAdded_WithSources(t *testing.T) {
	s1, err := open("../data/checker/api_security_global_added_revision.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_added_base.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	anonymous := requireSingleChange(t, errs, checker.APIGlobalSecurityAnonymousAccessAddedId)
	require.Equal(t, "../data/checker/api_security_global_added_revision.yaml", anonymous.GetBaseSource().File)
	require.Equal(t, 5, anonymous.GetBaseSource().Line)
}

// removing an endpoint security, with source tracking: the change points at the
// operation's "security" field of the base spec, and at the operation in the
// revision, which has no "security" field.
func TestAPISecurityDeleted_WithSources(t *testing.T) {
	s1, err := open("../data/checker/api_security_added_revision.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_added_base.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.APISecurityAnonymousAccessAddedId)

	require.NotEmpty(t, errs[0].GetBaseSource())
	require.Equal(t, "../data/checker/api_security_added_revision.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 23, errs[0].GetBaseSource().Line)
	require.Equal(t, "../data/checker/api_security_added_base.yaml", errs[0].GetRevisionSource().File)
}

type securityFinding struct {
	Id    string
	Level checker.Level
	Path  string
}

func securityFindings(t *testing.T, base, revision string, opts ...checker.Option) []securityFinding {
	t.Helper()
	result := []securityFinding{}
	for _, change := range securityChanges(t, "api_security_anonymous_"+base, "api_security_anonymous_"+revision, opts...) {
		result = append(result, securityFinding{Id: change.GetId(), Level: change.GetLevel(), Path: change.GetPath()})
	}
	return result
}

func securityChanges(t *testing.T, base, revision string, opts ...checker.Option) checker.Changes {
	t.Helper()
	s1, err := open("../data/checker/" + base + ".yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/" + revision + ".yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck, opts...), d, osm, checker.INFO)
}

// No security requirement means anonymous access, so the first requirement
// narrows access and removing the last one widens it. Each fixture has a
// /pets operation that varies and a /health operation that always declares
// `security: []`, so it never inherits the global requirement.
func TestAPISecurityAnonymousAccess(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		revision string
		want     []securityFinding
	}{
		{
			name: "operation secured", base: "none", revision: "op_api_key",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
				{checker.APISecurityAddedCheckId, checker.INFO, "/pets"},
			},
		},
		{
			name: "operation made anonymous", base: "op_api_key", revision: "none",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
			},
		},
		{
			name: "global security added to anonymous api", base: "none", revision: "global_api_key",
			want: []securityFinding{
				{checker.APIGlobalSecurityAnonymousAccessRemovedId, checker.ERR, ""},
				{checker.APIGlobalSecurityAddedCheckId, checker.INFO, ""},
			},
		},
		{
			name: "global security removed", base: "global_api_key", revision: "none",
			want: []securityFinding{
				{checker.APIGlobalSecurityAnonymousAccessAddedId, checker.INFO, ""},
			},
		},
		{
			name: "operation overrides global security with an empty list", base: "global_api_key", revision: "global_api_key_op_anonymous",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
			},
		},
		{
			name: "operation drops its empty list and inherits global security", base: "global_api_key_op_anonymous", revision: "global_api_key",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
				{checker.APISecurityAddedCheckId, checker.INFO, "/pets"},
			},
		},
		{
			name: "explicit anonymous alternative removed", base: "op_api_key_or_anonymous", revision: "op_api_key",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
			},
		},
		{
			name: "explicit anonymous alternative added", base: "op_api_key", revision: "op_api_key_or_anonymous",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
			},
		},
		{
			name: "operation replaces inherited global security with its own", base: "global_api_key", revision: "global_api_key_op_bearer",
			want: []securityFinding{
				{checker.APISecurityRemovedCheckId, checker.ERR, "/pets"},
				{checker.APISecurityAddedCheckId, checker.INFO, "/pets"},
			},
		},
		{
			name: "global security added while every operation has its own", base: "op_api_key", revision: "global_api_key_op_api_key",
			want: []securityFinding{
				{checker.APIGlobalSecurityAddedCheckId, checker.INFO, ""},
			},
		},
		{
			name: "global security added while every operation declares an empty list", base: "op_anonymous", revision: "global_api_key_op_anonymous",
			want: []securityFinding{
				{checker.APIGlobalSecurityAddedCheckId, checker.INFO, ""},
			},
		},
		{
			name: "global security added while the operation gets its own", base: "none", revision: "global_api_key_op_bearer",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
				{checker.APISecurityAddedCheckId, checker.INFO, "/pets"},
				{checker.APIGlobalSecurityAddedCheckId, checker.INFO, ""},
			},
		},
		{
			name: "empty list declared where there was nothing to inherit", base: "none", revision: "op_anonymous",
			want: []securityFinding{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.want, securityFindings(t, tc.base, tc.revision))
		})
	}
}

// An operation below the stability threshold is out of scope even when only
// the security it inherits shows the change.
func TestAPISecurityAnonymousAccessStability(t *testing.T) {
	require.Empty(t, securityFindings(t, "alpha_global_api_key", "alpha_global_api_key_op_anonymous"))
	require.Equal(t, []securityFinding{
		{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
	}, securityFindings(t, "alpha_global_api_key", "alpha_global_api_key_op_anonymous", checker.WithStabilityLevel(checker.StabilityAlpha)))
}

func TestAPISecurityAnonymousAccessText(t *testing.T) {
	l := checker.NewDefaultLocalizer()
	require.Equal(t, "the endpoint no longer allows anonymous access", l(checker.APISecurityAnonymousAccessRemovedId))
	require.Equal(t, "the endpoint now allows anonymous access", l(checker.APISecurityAnonymousAccessAddedId))
	require.Equal(t, "the API no longer allows anonymous access to endpoints without their own security requirements", l(checker.APIGlobalSecurityAnonymousAccessRemovedId))
	require.Equal(t, "the API now allows anonymous access to endpoints without their own security requirements", l(checker.APIGlobalSecurityAnonymousAccessAddedId))
}

// The same change to a security list is reported the same way whether the list
// is the top-level one or an operation's own, apart from the rule id prefix.
func TestAPISecurityListPositions(t *testing.T) {
	lists := map[string]openapi3.SecurityRequirements{
		"none":                 {},
		"api key":              {{"apiKey": {}}},
		"api key or anyone":    {{"apiKey": {}}, {}},
		"oauth read":           {{"oauth": {"read"}}},
		"oauth read or anyone": {{"oauth": {"read"}}, {}},
		"oauth read write":     {{"oauth": {"read", "write"}}},
	}

	tests := []struct {
		base, revision string
		want           []securityFinding
	}{
		{"api key", "api key or anyone", []securityFinding{
			{checker.APISecurityAnonymousAccessAddedId, checker.INFO, ""},
		}},
		{"api key or anyone", "api key", []securityFinding{
			{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, ""},
		}},
		{"api key", "none", []securityFinding{
			{checker.APISecurityAnonymousAccessAddedId, checker.INFO, ""},
		}},
		{"none", "api key", []securityFinding{
			{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, ""},
			{checker.APISecurityAddedCheckId, checker.INFO, ""},
		}},
		{"oauth read or anyone", "oauth read write", []securityFinding{
			{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, ""},
			{checker.APISecurityScopeAddedId, checker.ERR, ""},
		}},
		{"oauth read write", "oauth read or anyone", []securityFinding{
			{checker.APISecurityAnonymousAccessAddedId, checker.INFO, ""},
			{checker.APISecurityScopeRemovedId, checker.INFO, ""},
		}},
	}

	// findings returns the changes of moving /pets from base to revision,
	// either through the top-level list it inherits or through its own.
	findings := func(t *testing.T, base, revision openapi3.SecurityRequirements, global bool) []securityFinding {
		t.Helper()
		spec := func(security openapi3.SecurityRequirements) *load.SpecInfo {
			s, err := open("../data/checker/api_security_anonymous_none.yaml")
			require.NoError(t, err)
			if global {
				s.Spec.Security = security
			} else {
				s.Spec.Paths.Value("/pets").Get.Security = &security
			}
			return s
		}
		d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), spec(base), spec(revision))
		require.NoError(t, err)
		result := []securityFinding{}
		for _, change := range checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO) {
			id := strings.Replace(change.GetId(), "api-global-security-", "api-security-", 1)
			result = append(result, securityFinding{Id: id, Level: change.GetLevel()})
		}
		return result
	}

	for _, tc := range tests {
		t.Run(tc.base+" to "+tc.revision, func(t *testing.T) {
			base, revision := lists[tc.base], lists[tc.revision]
			require.ElementsMatch(t, tc.want, findings(t, base, revision, false), "operation")
			require.ElementsMatch(t, tc.want, findings(t, base, revision, true), "top level")
		})
	}
}
