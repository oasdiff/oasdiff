package checker_test

import (
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

var changes = checker.Changes{
	checker.ApiChange{
		Id:        "api-deleted",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/test",
	},
	checker.ApiChange{
		Id:        "api-added",
		Level:     checker.INFO,
		Operation: "GET",
		Path:      "/test",
	},
	checker.ComponentChange{
		Id:    "component-added",
		Level: checker.INFO,
	},
	checker.SecurityChange{
		Id:    "security-added",
		Level: checker.INFO,
	},
}

func TestChanges_Sort(t *testing.T) {
	slices.SortFunc(changes, checker.CompareChanges)
}

func TestChanges_IsBreaking(t *testing.T) {
	for _, c := range changes {
		require.True(t, c.IsBreaking() == (c.GetLevel() != checker.INFO))
	}
}

func TestChanges_Count(t *testing.T) {
	lc := changes.GetLevelCount()
	require.Equal(t, 3, lc[checker.INFO])
	require.Equal(t, 0, lc[checker.WARN])
	require.Equal(t, 1, lc[checker.ERR])
}

func TestIsEmpty_EmptyIncludeWarns(t *testing.T) {
	bcErrors := checker.Changes{}
	require.False(t, bcErrors.HasLevelOrHigher(checker.WARN))
}

func TestIsEmpty_EmptyExcludeWarns(t *testing.T) {
	bcErrors := checker.Changes{}
	require.False(t, bcErrors.HasLevelOrHigher(checker.ERR))
}

func TestIsEmpty_OneErrIncludeWarns(t *testing.T) {
	bcErrors := checker.Changes{
		checker.ApiChange{Level: checker.ERR},
	}
	require.True(t, bcErrors.HasLevelOrHigher(checker.WARN))
}

func TestIsEmpty_OneErrExcludeWarns(t *testing.T) {
	bcErrors := checker.Changes{
		checker.ApiChange{Level: checker.ERR},
	}
	require.True(t, bcErrors.HasLevelOrHigher(checker.ERR))
}

func TestIsEmpty_OneWarnIncludeWarns(t *testing.T) {
	bcErrors := checker.Changes{
		checker.ApiChange{Level: checker.WARN},
	}
	require.True(t, bcErrors.HasLevelOrHigher(checker.WARN))
}

func TestIsEmpty_OneWarnExcludeWarns(t *testing.T) {
	bcErrors := checker.Changes{
		checker.ApiChange{Level: checker.WARN},
	}
	require.False(t, bcErrors.HasLevelOrHigher(checker.ERR))
}

func TestCompareChanges_ByArgsLength(t *testing.T) {
	a := checker.ApiChange{
		Id:        "test",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/test",
		Args:      []any{"a"},
	}
	b := checker.ApiChange{
		Id:        "test",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/test",
		Args:      []any{"a", "b"},
	}
	result := checker.CompareChanges(a, b)
	require.Less(t, result, 0)
}

func TestCompareChanges_ByArgsValue(t *testing.T) {
	a := checker.ApiChange{
		Id:        "test",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/test",
		Args:      []any{"a"},
	}
	b := checker.ApiChange{
		Id:        "test",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/test",
		Args:      []any{"b"},
	}
	result := checker.CompareChanges(a, b)
	require.Less(t, result, 0)
}

func TestCompareChanges_Equal(t *testing.T) {
	a := checker.ApiChange{
		Id:        "test",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/test",
		Args:      []any{"a"},
	}
	b := checker.ApiChange{
		Id:        "test",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/test",
		Args:      []any{"a"},
	}
	result := checker.CompareChanges(a, b)
	require.Equal(t, 0, result)
}
