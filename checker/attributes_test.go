package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestBreaking_Attributes(t *testing.T) {
	s1, err := open("../data/attributes/base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/attributes/revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(checker.WithAttributes([]string{"x-test"})), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)

	require.Equal(t, map[string]any{"x-test": []any{"xyz", float64(456)}}, errs[0].GetAttributes())
	require.Equal(t, map[string]any{"x-test": "abc"}, errs[1].GetAttributes())
}

func TestBreaking_AttributesNone(t *testing.T) {
	s1, err := open("../data/attributes/base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/attributes/revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(checker.WithAttributes([]string{"x-other"})), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)

	require.Empty(t, errs[0].GetAttributes())
	require.Empty(t, errs[1].GetAttributes())
}

func TestBreaking_AttributesReverse(t *testing.T) {
	s1, err := open("../data/attributes/revision.yaml")
	require.NoError(t, err)

	s2, err := open("../data/attributes/base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(checker.WithAttributes([]string{"x-test"})), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)

	require.Equal(t, map[string]any{"x-test": []any{float64(123), float64(456)}}, errs[0].GetAttributes())
	require.Equal(t, map[string]any{"x-test": float64(123)}, errs[1].GetAttributes())
}

func TestBreaking_AttributesTwo(t *testing.T) {
	s1, err := open("../data/attributes/base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/attributes/revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(checker.WithAttributes([]string{"x-test", "x-test2"})), d, osm)
	require.Len(t, errs, 2)

	require.Equal(t, map[string]any{"x-test": []any{"xyz", float64(456)}}, errs[0].GetAttributes())
	require.Equal(t, map[string]any{"x-test": "abc", "x-test2": "def"}, errs[1].GetAttributes())
}
