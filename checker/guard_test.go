package checker_test

import (
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func guardChanges(t *testing.T, config *checker.Config) checker.Changes {
	t.Helper()
	return changesBetween(t, config, "../data/checker/read_only_guard_base.yaml", "../data/checker/read_only_guard_revision.yaml")
}

func changesBetween(t *testing.T, config *checker.Config, base, revision string) checker.Changes {
	t.Helper()
	s1, err := open(base)
	require.NoError(t, err)
	s2, err := open(revision)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	return checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
}

// findChangeByText returns the first change with the given id whose rendered
// text contains substr, distinguishing changes that share an id.
func findChangeByText(t *testing.T, changes checker.Changes, id, substr string) checker.Change {
	t.Helper()
	for _, c := range changes {
		if c.GetId() == id && strings.Contains(c.GetUncolorizedText(checker.NewDefaultLocalizer()), substr) {
			return c
		}
	}
	t.Fatalf("no change with id %q and text containing %q", id, substr)
	return nil
}

// A pattern added to a read-only request property is reported at info with a
// comment naming the reason, while the same edit on an ordinary sibling
// property keeps the rule's level. The check has no read-only logic of its
// own; levelByGuards derives the verdict.
func TestGuard_ReadOnlyRequestProperty(t *testing.T) {
	changes := guardChanges(t, allChecksConfig())

	readOnly := findChangeByText(t, changes, checker.RequestPropertyPatternAddedId, "`id`")
	require.Equal(t, checker.INFO, readOnly.GetLevel())
	require.Contains(t, readOnly.GetComment(checker.NewDefaultLocalizer()), "read-only")

	plain := findChangeByText(t, changes, checker.RequestPropertyPatternAddedId, "`name`")
	require.Equal(t, checker.ERR, plain.GetLevel())
	require.NotContains(t, plain.GetComment(checker.NewDefaultLocalizer()), "read-only")
}

// The write-only mirror: a type change on a write-only response property is
// nullified because the property never appears in responses, even though the
// change is otherwise incomparable and an error. The same property in the
// request body keeps the error: write-only protects only the response side.
func TestGuard_WriteOnlyResponseProperty(t *testing.T) {
	changes := guardChanges(t, allChecksConfig())

	writeOnly := findChangeByText(t, changes, checker.ResponsePropertyTypeChangedId, "`secret`")
	require.Equal(t, checker.INFO, writeOnly.GetLevel())
	require.Contains(t, writeOnly.GetComment(checker.NewDefaultLocalizer()), "write-only")

	request := findChangeByText(t, changes, checker.RequestPropertyTypeChangedId, "`secret`")
	require.Equal(t, checker.ERR, request.GetLevel())
	require.NotContains(t, request.GetComment(checker.NewDefaultLocalizer()), "write-only")
}

// An explicit --severity-levels entry claims the whole id: the guard neither
// lowers the overridden level nor attaches its comment, mirroring
// capByDisclaimers.
func TestGuard_OverriddenLevelWins(t *testing.T) {
	changes := guardChanges(t, allChecksConfig(checker.WithSeverityLevels(map[string]checker.Level{
		checker.RequestPropertyPatternAddedId: checker.WARN,
	})))

	readOnly := findChangeByText(t, changes, checker.RequestPropertyPatternAddedId, "`id`")
	require.Equal(t, checker.WARN, readOnly.GetLevel())
	require.NotContains(t, readOnly.GetComment(checker.NewDefaultLocalizer()), "read-only")
}

const negatedComment = "under `not`"

// The fixture raises the minLength of two shared schemas from 5 to 10. Under
// one `not` the change accepts more values, so it is safe in a request and
// breaking in a response, the reverse of its rule's level. A schema used both
// plainly and under `not` is reported once for each, and under two `not`s the
// change keeps its rule's level.
func TestGuard_Negated(t *testing.T) {
	changes := changesBetween(t, allChecksConfig(), "../data/checker/negated_base.yaml", "../data/checker/negated_revision.yaml")
	l := checker.NewDefaultLocalizer()

	plain := findChangeByText(t, changes, checker.RequestPropertyMinLengthIncreasedId, "`plain`")
	require.Equal(t, checker.ERR, plain.GetLevel())
	require.NotContains(t, plain.GetComment(l), negatedComment)

	negated := findChangeByText(t, changes, checker.RequestPropertyMinLengthIncreasedId, "`excluded/not`")
	require.Equal(t, checker.INFO, negated.GetLevel())
	require.Contains(t, negated.GetComment(l), negatedComment)

	twice := findChangeByText(t, changes, checker.RequestPropertyMinLengthIncreasedId, "`twice/not/not`")
	require.Equal(t, checker.ERR, twice.GetLevel())
	require.NotContains(t, twice.GetComment(l), negatedComment)

	response := findChangeByText(t, changes, "response-property-min-length-increased", "`excluded/not`")
	require.Equal(t, checker.ERR, response.GetLevel())
	require.Contains(t, response.GetComment(l), negatedComment)
}

// The opposite edit: lowering minLength under `not` accepts fewer values, so
// a change whose rule is info in a request is reported as an error.
func TestGuard_NegatedRaisesLevel(t *testing.T) {
	changes := changesBetween(t, allChecksConfig(), "../data/checker/negated_revision.yaml", "../data/checker/negated_base.yaml")
	l := checker.NewDefaultLocalizer()

	plain := findChangeByText(t, changes, checker.RequestPropertyMinLengthDecreasedId, "`plain`")
	require.Equal(t, checker.INFO, plain.GetLevel())

	negated := findChangeByText(t, changes, checker.RequestPropertyMinLengthDecreasedId, "`excluded/not`")
	require.Equal(t, checker.ERR, negated.GetLevel())
	require.Contains(t, negated.GetComment(l), negatedComment)
}

// A level set with --severity-levels is kept under `not` too.
func TestGuard_NegatedOverriddenLevelWins(t *testing.T) {
	changes := changesBetween(t, allChecksConfig(checker.WithSeverityLevels(map[string]checker.Level{
		checker.RequestPropertyMinLengthIncreasedId: checker.WARN,
	})), "../data/checker/negated_base.yaml", "../data/checker/negated_revision.yaml")

	negated := findChangeByText(t, changes, checker.RequestPropertyMinLengthIncreasedId, "`excluded/not`")
	require.Equal(t, checker.WARN, negated.GetLevel())
	require.NotContains(t, negated.GetComment(checker.NewDefaultLocalizer()), negatedComment)
}
