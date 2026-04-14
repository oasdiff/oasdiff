package checker_test

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func loadFrom(t *testing.T, prefix string, v int) *load.SpecInfo {
	t.Helper()

	path := fmt.Sprintf(prefix+"spec%d.yaml", v)
	loader := openapi3.NewLoader()
	oas, err := loader.LoadFromFile(path)
	require.NoError(t, err)
	return &load.SpecInfo{Spec: oas, Url: path}
}

func TestComposed_Empty(t *testing.T) {
	s1 := []*load.SpecInfo{
		loadFrom(t, "../data/composed/base/", 1),
	}

	s2 := []*load.SpecInfo{
		loadFrom(t, "../data/composed/base/", 1),
	}

	diffReport, _, err := diff.GetPathsDiff(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Nil(t, diffReport)
}

func TestComposed_Duplicate(t *testing.T) {
	s1 := []*load.SpecInfo{
		loadFrom(t, "../data/composed/base/", 1),
		loadFrom(t, "../data/composed/base/", 1),
	}

	s2 := []*load.SpecInfo{}

	config := diff.NewConfig()
	_, _, err := diff.GetPathsDiff(config, s1, s2)
	require.Error(t, err)
}

func TestComposed_Issue500(t *testing.T) {
	s1 := []*load.SpecInfo{
		loadFrom(t, "../data/composed/issue500/", 1),
		loadFrom(t, "../data/composed/issue500/", 2),
	}

	config := diff.NewConfig()
	_, _, err := diff.GetPathsDiff(config, s1, s1)
	require.NoError(t, err)
}

func TestComposed_CompareMostRecent(t *testing.T) {
	s1 := []*load.SpecInfo{
		loadFrom(t, "../data/composed/base/", 1),
		loadFrom(t, "../data/composed/base/", 2),
	}

	s2 := []*load.SpecInfo{
		loadFrom(t, "../data/composed/revision/", 1),
		loadFrom(t, "../data/composed/revision/", 2),
	}

	diffReport, _, err := diff.GetPathsDiff(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Nil(t, diffReport)
}
