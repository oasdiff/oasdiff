package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// detecting deprecated request properties with sunset date
func TestRequestPropertyDeprecationCheck(t *testing.T) {
	s1, err := open("../data/deprecation/request_property_deprecation_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/deprecation/request_property_deprecation_spec.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.RequestPropertyDeprecationCheck(d, osm, config)

	found := false
	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyDeprecatedId {
			found = true
			t.Logf("Found deprecated request property: %+v", c)
		}
	}
	if !found {
		t.Errorf("Expected RequestPropertyDeprecatedId in changes, got: %+v", changes)
	}
}

// detecting deprecated request properties in allOf schemas
func TestRequestPropertyDeprecationCheck_AllOf(t *testing.T) {
	s1, err := open("../data/deprecation/request_property_deprecation_allof_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/deprecation/request_property_deprecation_allof_spec.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.RequestPropertyDeprecationCheck(d, osm, config)

	found := false
	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyDeprecatedId {
			found = true
			t.Logf("Found deprecated request allOf property: %+v", c)
		}
	}
	if !found {
		t.Errorf("Expected RequestPropertyDeprecatedId in changes, got: %+v", changes)
	}
}

// Ensuring no duplicate deprecation reports for the same property
func TestRequestPropertyDeprecationCheck_NoDuplicates(t *testing.T) {
	s1, err := open("../data/deprecation/request_property_deprecation_allof_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/deprecation/request_property_deprecation_allof_spec.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.RequestPropertyDeprecationCheck(d, osm, config)

	// Count occurrences of each property
	propCount := make(map[string]int)
	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyDeprecatedId {
			propCount[c.GetText(checker.NewDefaultLocalizer())]++
		}
	}

	// Each property should only appear once
	for prop, count := range propCount {
		if count > 1 {
			t.Errorf("Property %s appears %d times, expected 1", prop, count)
		}
	}
}
