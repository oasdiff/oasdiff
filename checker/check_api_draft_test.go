package checker_test

import (
	"fmt"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func getDraftFile(file string) string {
	return fmt.Sprintf("../data/draft/%s", file)
}

func draftChecksConfig() *checker.Config {
	config := checker.NewConfig(checker.GetAllChecks())
	config.IncludeStabilityLevels = map[string]bool{
		"draft": true,
		"alpha": true,
	}
	return config
}

func noDraftChecksConfig() *checker.Config {
	return checker.NewConfig(checker.GetAllChecks())
}

// Test: endpoint marked as draft is detected when IncludeStabilityLevels includes draft
func TestEndpointDraft_Detected(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_endpoint_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	// Should detect endpoint-draft for GET and POST
	draftChanges := filterById(errs, checker.EndpointDraftId)
	require.Len(t, draftChanges, 2, "expected 2 endpoint-draft changes (GET + POST)")

	for _, c := range draftChanges {
		require.IsType(t, checker.ApiChange{}, c)
		e := c.(checker.ApiChange)
		require.Equal(t, checker.EndpointDraftId, e.Id)
		require.Equal(t, "/api/test", e.Path)
	}
}

// Test: endpoint marked as draft is NOT detected when IncludeStabilityLevels is not set
func TestEndpointDraft_NotDetectedWithoutFlag(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_endpoint_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(noDraftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.EndpointDraftId)
	require.Len(t, draftChanges, 0, "expected no endpoint-draft changes when flag is not set")
}

// Test: endpoint already draft in base and revision does NOT produce a change
func TestEndpointDraft_AlreadyDraftNoChange(t *testing.T) {
	s1, err := open(getDraftFile("base_already_draft.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_property_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.EndpointDraftId)
	require.Len(t, draftChanges, 0, "expected no endpoint-draft changes when endpoint was already draft")
}

// Test: newly added path with draft endpoint is detected
func TestEndpointDraft_NewPathDetected(t *testing.T) {
	s1, err := open(getDraftFile("base_simple.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_new_path_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.EndpointDraftId)
	require.Len(t, draftChanges, 1, "expected 1 endpoint-draft change for newly added draft path")

	e := draftChanges[0].(checker.ApiChange)
	require.Equal(t, "/api/new-draft", e.Path)
	require.Equal(t, "GET", e.Operation)
}

// Test: request property marked as draft is detected when IncludeStabilityLevels includes draft
func TestRequestPropertyDraft_Detected(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_property_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.RequestPropertyDraftId)
	require.Len(t, draftChanges, 1, "expected 1 request-property-draft change")

	e := draftChanges[0].(checker.ApiChange)
	require.Equal(t, checker.RequestPropertyDraftId, e.Id)
	require.Equal(t, "POST", e.Operation)
	require.Equal(t, "/api/test", e.Path)
}

// Test: request property marked as draft is NOT detected when flag is not set
func TestRequestPropertyDraft_NotDetectedWithoutFlag(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_property_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(noDraftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.RequestPropertyDraftId)
	require.Len(t, draftChanges, 0, "expected no request-property-draft changes when flag is not set")
}

// Test: request property already draft in base does NOT produce a change
func TestRequestPropertyDraft_AlreadyDraftNoChange(t *testing.T) {
	s1, err := open(getDraftFile("base_already_draft.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_property_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.RequestPropertyDraftId)
	require.Len(t, draftChanges, 0, "expected no request-property-draft changes when property was already draft")
}

// Test: response property marked as draft is detected when IncludeStabilityLevels includes draft
func TestResponsePropertyDraft_Detected(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_property_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.ResponsePropertyDraftId)
	require.Len(t, draftChanges, 1, "expected 1 response-property-draft change")

	e := draftChanges[0].(checker.ApiChange)
	require.Equal(t, checker.ResponsePropertyDraftId, e.Id)
	require.Equal(t, "POST", e.Operation)
	require.Equal(t, "/api/test", e.Path)
}

// Test: response property marked as draft is NOT detected when flag is not set
func TestResponsePropertyDraft_NotDetectedWithoutFlag(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_property_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(noDraftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.ResponsePropertyDraftId)
	require.Len(t, draftChanges, 0, "expected no response-property-draft changes when flag is not set")
}

// Test: response property already draft in base does NOT produce a change
func TestResponsePropertyDraft_AlreadyDraftNoChange(t *testing.T) {
	s1, err := open(getDraftFile("base_already_draft.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_property_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.ResponsePropertyDraftId)
	require.Len(t, draftChanges, 0, "expected no response-property-draft changes when property was already draft")
}

// Test: no draft changes when base and revision are identical
func TestDraft_NoDiff(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.EndpointDraftId)
	require.Len(t, draftChanges, 0, "expected no endpoint-draft changes when specs are identical")

	draftChanges = filterById(errs, checker.RequestPropertyDraftId)
	require.Len(t, draftChanges, 0, "expected no request-property-draft changes when specs are identical")

	draftChanges = filterById(errs, checker.ResponsePropertyDraftId)
	require.Len(t, draftChanges, 0, "expected no response-property-draft changes when specs are identical")
}

// Test: nil config returns no draft changes
func TestDraft_NilConfig(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_endpoint_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	result := checker.APIDraftCheck(d, osm, nil)
	require.Len(t, result, 0, "expected no changes with nil config")

	result = checker.RequestPropertyDraftCheck(d, osm, nil)
	require.Len(t, result, 0, "expected no changes with nil config")

	result = checker.ResponsePropertyDraftCheck(d, osm, nil)
	require.Len(t, result, 0, "expected no changes with nil config")
}

// Test: draft changes have correct message text
func TestEndpointDraft_MessageText(t *testing.T) {
	s1, err := open(getDraftFile("base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getDraftFile("revision_endpoint_draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(draftChecksConfig(), d, osm, checker.INFO)

	draftChanges := filterById(errs, checker.EndpointDraftId)
	require.NotEmpty(t, draftChanges)

	msg := draftChanges[0].GetUncolorizedText(checker.NewDefaultLocalizer())
	require.Contains(t, msg, "draft", "message should contain the word 'draft'")
}

// Helper: filter changes by ID
func filterById(changes checker.Changes, id string) checker.Changes {
	result := make(checker.Changes, 0)
	for _, c := range changes {
		if c.GetId() == id {
			result = append(result, c)
		}
	}
	return result
}
