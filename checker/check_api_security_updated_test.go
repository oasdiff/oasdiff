package checker_test

import (
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
	removed := requireSingleChange(t, errs, checker.APISecurityAnonymousAccessRemovedId)
	require.Equal(t, "POST", removed.GetOperation())
	require.Equal(t, "/subscribe", removed.GetPath())
	require.Equal(t, "the endpoint no longer allows anonymous access and now only accepts `petstore_auth: [read:pets, write:pets]`", removed.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// adding a global security that no endpoint inherits
func TestAPIGlobalSecurityAddedNotInherited(t *testing.T) {
	errs := checkChanges(t, checker.APISecurityUpdatedCheck, "../data/checker/api_security_anonymous_op_api_key.yaml", "../data/checker/api_security_anonymous_global_api_key_op_api_key.yaml")
	added := requireSingleChange(t, errs, checker.APIGlobalSecurityAddedCheckId)
	require.Equal(t, "the security scheme `apiKey` was added to the API", added.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing a global security from the API
func TestAPIGlobalSecurityyDeleted(t *testing.T) {
	s1, err := open("../data/checker/api_security_global_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	added := requireSingleChange(t, errs, checker.APISecurityAnonymousAccessAddedId)
	require.Equal(t, "/subscribe", added.GetPath())
}

// removing a global security that no endpoint inherits
func TestAPIGlobalSecurityRemovedNotInherited(t *testing.T) {
	errs := checkChanges(t, checker.APISecurityUpdatedCheck, "../data/checker/api_security_anonymous_global_api_key_op_api_key.yaml", "../data/checker/api_security_anonymous_op_api_key.yaml")
	removed := requireSingleChange(t, errs, checker.APIGlobalSecurityRemovedCheckId)
	require.Equal(t, "the security scheme `apiKey` was removed from the API", removed.GetUncolorizedText(checker.NewDefaultLocalizer()))
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
	removed := requireSingleChange(t, errs, checker.APISecurityScopeRemovedId)
	require.Equal(t, checker.INFO, removed.GetLevel())
	require.Equal(t, "/subscribe", removed.GetPath())
	require.Equal(t, checker.APISecurityInheritedCommentId, removed.(checker.ApiChange).Comment)
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
	added := requireSingleChange(t, errs, checker.APISecurityScopeAddedId)
	require.Equal(t, checker.ERR, added.GetLevel())
	require.Equal(t, "/subscribe", added.GetPath())
}

// replacing the top-level requirement rejects clients of every operation that
// inherits it
func TestAPIGlobalSecurityReplaced(t *testing.T) {
	require.ElementsMatch(t, []securityFinding{
		{checker.APISecurityRemovedCheckId, checker.ERR, "/pets"},
		{checker.APISecurityAddedCheckId, checker.INFO, "/pets"},
	}, securityFindings(t, "global_api_key", "global_bearer"))
}

// a top-level change that no operation inherits affects no client
func TestAPIGlobalSecurityNotInherited(t *testing.T) {
	errs := checkChanges(t, checker.APISecurityUpdatedCheck, "../data/checker/api_security_anonymous_op_api_key.yaml", "../data/checker/api_security_anonymous_global_api_key_op_api_key.yaml")
	added := requireSingleChange(t, errs, checker.APIGlobalSecurityAddedCheckId)
	require.Equal(t, checker.INFO, added.GetLevel())
	require.Equal(t, "No endpoint inherits the API's top-level security requirements, so this change affects no client.", added.GetComment(checker.NewDefaultLocalizer()))
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
	removed := requireSingleChange(t, errs, checker.APISecurityAnonymousAccessRemovedId)
	requireApiChange(t, checker.ApiChange{
		Id:        checker.APISecurityAnonymousAccessRemovedId,
		Args:      []any{[]string{"petstore_auth: [read:pets, write:pets]"}},
		Operation: "POST",
		Path:      "/subscribe",
		Source:    load.NewSource("../data/checker/api_security_added_revision.yaml"),
	}, removed)
	require.Equal(t, "the endpoint no longer allows anonymous access and now only accepts `petstore_auth: [read:pets, write:pets]`", removed.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// adding an alternative to an endpoint that already requires security
func TestAPISecurityAddedToSecuredEndpoint(t *testing.T) {
	errs := checkChanges(t, checker.APISecurityUpdatedCheck, "../data/checker/api_security_anonymous_global_api_key.yaml", "../data/checker/api_security_anonymous_global_api_key_op_bearer.yaml")
	added := requireChange(t, errs, checker.APISecurityAddedCheckId)
	require.Equal(t, "the endpoint scheme security `bearer` was added to the API", added.GetUncolorizedText(checker.NewDefaultLocalizer()))
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
	s1, err := open("../data/checker/api_security_global_added_revision.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_global_added_base.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISecurityUpdatedCheck), d, osm, checker.INFO)
	anonymous := requireSingleChange(t, errs, checker.APISecurityAnonymousAccessAddedId)
	require.NotEmpty(t, anonymous.GetBaseSource())
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
				{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
			},
		},
		{
			name: "global security removed", base: "global_api_key", revision: "none",
			want: []securityFinding{
				{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
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
		// The reverse change reports the opposite of each finding, so a
		// finding with no counterpart in one direction fails here.
		t.Run(tc.name+" reversed", func(t *testing.T) {
			want := []string{}
			for _, finding := range tc.want {
				want = append(want, oppositeSecurityId[finding.Id])
			}
			require.ElementsMatch(t, want, securityFindingIds(t, tc.revision, tc.base))
		})
	}
}

var oppositeSecurityId = map[string]string{
	checker.APISecurityAddedCheckId:             checker.APISecurityRemovedCheckId,
	checker.APISecurityRemovedCheckId:           checker.APISecurityAddedCheckId,
	checker.APIGlobalSecurityAddedCheckId:       checker.APIGlobalSecurityRemovedCheckId,
	checker.APIGlobalSecurityRemovedCheckId:     checker.APIGlobalSecurityAddedCheckId,
	checker.APISecurityAnonymousAccessAddedId:   checker.APISecurityAnonymousAccessRemovedId,
	checker.APISecurityAnonymousAccessRemovedId: checker.APISecurityAnonymousAccessAddedId,
}

// An operation below the stability threshold is out of scope even when only
// the security it inherits shows the change.
func TestAPISecurityAnonymousAccessStability(t *testing.T) {
	require.Empty(t, securityFindings(t, "alpha_global_api_key", "alpha_global_api_key_op_anonymous"))
	require.Equal(t, []securityFinding{
		{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
	}, securityFindings(t, "alpha_global_api_key", "alpha_global_api_key_op_anonymous", checker.WithStabilityLevel(checker.StabilityAlpha)))
}

func TestAPIInheritedSecurityAnonymousAccessStability(t *testing.T) {
	require.NotContains(t, securityFindingIds(t, "alpha_global_api_key", "alpha_global_anonymous"), checker.APISecurityAnonymousAccessAddedId)
	require.Contains(t, securityFindingIds(t, "alpha_global_api_key", "alpha_global_anonymous", checker.WithStabilityLevel(checker.StabilityAlpha)), checker.APISecurityAnonymousAccessAddedId)
}

// each alternative the endpoint now accepts is listed
func TestAPISecurityAnonymousAccessRemovedAlternatives(t *testing.T) {
	errs := checkChanges(t, checker.APISecurityUpdatedCheck, "../data/checker/api_security_anonymous_none.yaml", "../data/checker/api_security_anonymous_op_api_key_or_bearer.yaml")
	removed := requireSingleChange(t, errs, checker.APISecurityAnonymousAccessRemovedId)
	require.Equal(t, []any{[]string{"apiKey", "bearer"}}, removed.GetArgs())
	require.Equal(t, "the endpoint no longer allows anonymous access and now only accepts `apiKey, bearer`", removed.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func TestAPISecurityAnonymousAccessText(t *testing.T) {
	l := checker.NewDefaultLocalizer()
	require.Equal(t, "the endpoint no longer allows anonymous access and now only accepts apiKey", l(checker.APISecurityAnonymousAccessRemovedId, "apiKey"))
	require.Equal(t, "the endpoint now allows anonymous access", l(checker.APISecurityAnonymousAccessAddedId))
}

// A finding on an operation whose own definition did not change says the
// change came through the root list.
func TestAPISecurityInheritedComment(t *testing.T) {
	anonymous := func(base, revision string) string {
		t.Helper()
		changes := checkChanges(t, checker.APISecurityUpdatedCheck, "../data/checker/api_security_anonymous_"+base+".yaml", "../data/checker/api_security_anonymous_"+revision+".yaml")
		return requireSingleChange(t, changes, checker.APISecurityAnonymousAccessRemovedId).(checker.ApiChange).Comment
	}
	require.Equal(t, checker.APISecurityInheritedCommentId, anonymous("none", "global_api_key"))
	require.Empty(t, anonymous("none", "op_api_key"))
	require.Empty(t, anonymous("global_api_key_op_anonymous", "global_api_key"))
	require.Equal(t, "This endpoint inherits the API's top-level security requirements.", checker.NewDefaultLocalizer()(checker.APISecurityInheritedCommentId))
}

// The same change to a security list gives the same findings whether /pets
// declares the list itself or inherits it from the top level.
func TestAPISecurityListPositions(t *testing.T) {
	lists := map[string]openapi3.SecurityRequirements{
		"none":                 {},
		"api key":              {{"apiKey": {}}},
		"bearer":               {{"bearer": {}}},
		"api key or bearer":    {{"apiKey": {}}, {"bearer": {}}},
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
			{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
		}},
		{"api key or anyone", "api key", []securityFinding{
			{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
		}},
		{"api key", "none", []securityFinding{
			{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
		}},
		{"none", "api key", []securityFinding{
			{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
		}},
		{"api key", "bearer", []securityFinding{
			{checker.APISecurityRemovedCheckId, checker.ERR, "/pets"},
			{checker.APISecurityAddedCheckId, checker.INFO, "/pets"},
		}},
		{"api key", "api key or bearer", []securityFinding{
			{checker.APISecurityAddedCheckId, checker.INFO, "/pets"},
		}},
		{"api key or bearer", "api key", []securityFinding{
			{checker.APISecurityRemovedCheckId, checker.ERR, "/pets"},
		}},
		{"oauth read", "oauth read write", []securityFinding{
			{checker.APISecurityScopeAddedId, checker.ERR, "/pets"},
		}},
		{"oauth read write", "oauth read", []securityFinding{
			{checker.APISecurityScopeRemovedId, checker.INFO, "/pets"},
		}},
		{"oauth read or anyone", "oauth read write", []securityFinding{
			{checker.APISecurityAnonymousAccessRemovedId, checker.ERR, "/pets"},
		}},
		{"oauth read write", "oauth read or anyone", []securityFinding{
			{checker.APISecurityAnonymousAccessAddedId, checker.INFO, "/pets"},
		}},
	}

	findings := func(t *testing.T, base, revision openapi3.SecurityRequirements, topLevel bool) []securityFinding {
		t.Helper()
		spec := func(security openapi3.SecurityRequirements) *load.SpecInfo {
			s, err := open("../data/checker/api_security_anonymous_none.yaml")
			require.NoError(t, err)
			if topLevel {
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
			result = append(result, securityFinding{Id: change.GetId(), Level: change.GetLevel(), Path: change.GetPath()})
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
