package checker_test

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// A change compared branch by branch cannot be reported at error, whatever the
// rule declares, because a sibling branch may still guarantee what this one
// dropped. Here both branches require id and one drops it, so the merged
// schema is unchanged and the error would be a false positive.
func TestDisclaimer_AllOfCapsSeverity(t *testing.T) {
	change := singleChange(t, "../data/checker/disclaimer_allof_base.yaml", "../data/checker/disclaimer_allof_revision.yaml", checker.ResponsePropertyBecameOptionalId)

	require.Equal(t, checker.ERR, declaredLevel(t, change.GetId()), "the rule still declares the level it means when the whole schema is visible")
	require.Equal(t, checker.WARN, change.GetLevel(), "the disclaimer caps the change")
	require.Contains(t, change.GetComment(checker.NewDefaultLocalizer()), "--flatten-allof")
}

// Flattening removes the reason for the disclaimer, so the rule's own level
// applies again. Here the merged schemas are identical, so nothing is reported
// at all, which is the exact answer the capped warning stood in for.
func TestDisclaimer_FlatteningRemovesIt(t *testing.T) {
	// Flattening happens at load, which is why an allOf surviving into the
	// diff is what the disclaimer keys off.
	loader := openapi3.NewLoader()
	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/checker/disclaimer_allof_base.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)
	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/checker/disclaimer_allof_revision.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	require.Empty(t, checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO))
}

// A level the caller set explicitly is a decision about that rule, so a
// disclaimer does not lower it.
func TestDisclaimer_ExplicitLevelWins(t *testing.T) {
	s1, err := open("../data/checker/disclaimer_allof_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/disclaimer_allof_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(checker.GetAllChecks(),
		checker.WithSeverityLevels(map[string]checker.Level{checker.ResponsePropertyBecameOptionalId: checker.ERR}))
	changes := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	change := requireChange(t, changes, checker.ResponsePropertyBecameOptionalId)
	require.Equal(t, checker.ERR, change.GetLevel())
}

// Disclaimers reach a machine consumer as their names, so a script can tell a
// softened verdict from a confident one without matching on comment prose.
func TestDisclaimersSerializeAsNames(t *testing.T) {
	change := singleChange(t, "../data/checker/disclaimer_allof_base.yaml", "../data/checker/disclaimer_allof_revision.yaml", checker.ResponsePropertyBecameOptionalId)

	encoded, err := json.Marshal(formatters.NewChanges(checker.Changes{change}, checker.NewDefaultLocalizer()))
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"disclaimers":["all-of-not-flattened"]`)
}

// A change from an exact comparison carries none, and the field is omitted
// rather than serialized as an empty list.
func TestDisclaimersOmittedWhenNone(t *testing.T) {
	s1, err := open("../data/checker/request_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.NotEmpty(t, changes)

	encoded, err := json.Marshal(formatters.NewChanges(changes, checker.NewDefaultLocalizer()))
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "disclaimers")
}

// The allOf can be on an ancestor of the change rather than on the body, and
// the branches are still compared one at a time. Here only the property "p"
// composes, and a sibling branch still requires what one branch dropped.
func TestDisclaimer_AllOfOnAnAncestorProperty(t *testing.T) {
	change := singleChange(t, "../data/checker/disclaimer_nested_allof_base.yaml", "../data/checker/disclaimer_nested_allof_revision.yaml", checker.ResponsePropertyBecameOptionalId)

	require.Equal(t, checker.WARN, change.GetLevel())
}

// A property whose name happens to look like a composition path is not one.
func TestDisclaimer_PropertyNamedLikeAnAllOfPathIsNotOne(t *testing.T) {
	change := singleChange(t, "../data/checker/disclaimer_allof_lookalike_base.yaml", "../data/checker/disclaimer_allof_lookalike_revision.yaml", checker.ResponsePropertyBecameOptionalId)

	require.Equal(t, checker.ERR, change.GetLevel(), "nothing composes here, so nothing is uncertain")
}
