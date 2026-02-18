package report

import (
	"maps"
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

type DiffT interface {
	*diff.ExampleDiff |
		*diff.ServerDiff |
		*diff.ParameterDiff |
		*diff.VariableDiff |
		*diff.SchemaDiff |
		*diff.ResponseDiff |
		*diff.MediaTypeDiff |
		*diff.HeaderDiff |
		diff.SecurityScopesDiff |
		*diff.StringsDiff
}

func getKeys[diff DiffT](m map[string]diff) []string {
	return slices.Sorted(maps.Keys(m))
}
