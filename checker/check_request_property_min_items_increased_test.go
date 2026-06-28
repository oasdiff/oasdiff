package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// inclreasing request body min items is breaking
func TestRequestBodyMinItemsIncreased(t *testing.T) {
	s1, err := open("../data/checker/request_property_min_items_increased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_min_items_increased_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinItemsIncreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMinItemsIncreasedId,
		Args:        []any{uint64(20)},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_min_items_increased_revision.yaml"),
		OperationId: "addProduct",
	}, errs)
}

// descreasing request body min items is not breaking
func TestRequestBodyMinItemsDecreased(t *testing.T) {
	s1, err := open("../data/checker/request_property_min_items_increased_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_min_items_increased_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinItemsIncreasedCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}
