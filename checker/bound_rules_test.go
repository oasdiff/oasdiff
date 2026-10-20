package checker

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// boundSpecs and diff.SchemaBounds list the same keywords: every spec
// resolves to a bound, and every bound has a spec, so a keyword added to
// either side without the other fails instead of silently generating
// nothing.
func TestBoundSpecsMatchSchemaBounds(t *testing.T) {
	require.Len(t, boundRules(), 180)

	specs := map[string]bool{}
	for _, spec := range boundSpecs {
		specs[spec.keyword] = true
		_, ok := schemaBound(spec.keyword)
		require.True(t, ok, "%s does not resolve to a diff.SchemaBound", spec.keyword)
	}
	for _, bound := range diff.SchemaBounds {
		require.True(t, specs[bound.Keyword], "diff.SchemaBounds lists %s but boundSpecs does not; no rules are generated for it", bound.Keyword)
	}
}

// Hand-written coverage of a bound's set or unset must use the id boundRules
// skips by. The skip matches by id, so a hand-written rule covering such a
// cell under a differently formatted id would leave the cell to the generated
// check as well, and the edit would be reported twice.
func TestHandWrittenBoundIdsMatchTheGrammar(t *testing.T) {
	grammar := map[string]bool{}
	var cells []metaschema.Edit
	for _, spec := range boundSpecs {
		for _, direction := range []Direction{DirectionRequest, DirectionResponse} {
			for _, scope := range boundScopes(direction) {
				for _, action := range boundActions {
					if _, ok := boundEffect(spec.polarity, action); !ok {
						continue
					}
					cells = append(cells, metaschema.Edit{
						Location: boundLocation(direction, scope, spec.keyword),
						Action:   metaschema.Action(action.claim),
					})
					grammar[boundRuleId(direction, scope, spec.idName, action.verb)] = true
				}
			}
		}
	}

	for _, rule := range handWrittenRules() {
		if grammar[rule.Id] {
			continue
		}
		// a guarded rule is a variant of a cell fired only under its guard
		// (e.g. request-read-only-property-max-decreased); the cell's
		// unguarded owner is held to the grammar above
		if len(rule.Guards) > 0 {
			continue
		}
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				continue // the coverage audit reports invalid claims
			}
			for _, cell := range cells {
				if claim.Matches(cell) {
					t.Errorf("%s claims %s:%s, a cell boundRules generates; rename it to the generated id or the edit is reported twice",
						rule.Id, cell.Location, cell.Action)
				}
			}
		}
	}
}
