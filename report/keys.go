package report

import (
	"maps"
	"slices"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/utils"
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

func getKeys[diff DiffT](m map[string]diff) utils.StringList {
	return utils.StringList(slices.Sorted(maps.Keys(m)))
}
