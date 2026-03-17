package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
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

func TestChanges_Group(t *testing.T) {
	grouped := formatters.GroupChanges(changes, checker.NewDefaultLocalizer())
	require.Contains(t, grouped, formatters.ChangeGroup{Path: "/test", Operation: "GET"})
	require.Contains(t, grouped, formatters.ChangeGroup{Path: "components"})
	require.Contains(t, grouped, formatters.ChangeGroup{Path: "security"})
}
