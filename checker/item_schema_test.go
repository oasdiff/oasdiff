package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// OpenAPI 3.2 itemSchema types one item of a streamed body (SSE, JSON Lines,
// JSON sequences) rather than the body as a whole. The body-schema checks reach
// it through the media-type walkers, so a change to a streamed item is checked
// by the same rules, with an "(item schema)" detail telling the two apart.

func itemSchemaChanges(t *testing.T) checker.Changes {
	t.Helper()

	s1, err := open("../data/item-schema/base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/item-schema/revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	return checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)
}

// The headline case: adding a required property to a streamed request item
// breaks producers, and was reported as no change before itemSchema was diffed.
func TestItemSchema_RequestPropertyBecameRequiredIsBreaking(t *testing.T) {
	change := requireChange(t, itemSchemaChanges(t), checker.RequestPropertyBecameRequiredId)

	require.Equal(t, "/ingest", change.GetPath())
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Contains(t, change.GetUncolorizedText(checker.NewDefaultLocalizer()), "item schema")
}

// The response side reaches the item through the same walker.
func TestItemSchema_ResponsePropertyChangeIsReported(t *testing.T) {
	change := requireChange(t, itemSchemaChanges(t), checker.ResponsePropertyEnumValueRemovedId)

	require.Equal(t, "/events", change.GetPath())
	require.Contains(t, change.GetUncolorizedText(checker.NewDefaultLocalizer()), "item schema")
}

// An itemSchema appearing or disappearing is not a modification of an existing
// schema, so the existence check reports it, by the same contravariance as a
// body schema: constraining a request breaks producers, dropping the guarantee
// from a response breaks consumers.
func TestItemSchema_ExistenceIsClassifiedByContravariance(t *testing.T) {
	changes := itemSchemaChanges(t)

	added := requireChange(t, changes, checker.RequestBodyMediaTypeItemSchemaAddedId)
	require.Equal(t, "/drop", added.GetPath())
	require.Equal(t, checker.ERR, added.GetLevel())

	removed := requireChange(t, changes, checker.ResponseBodyMediaTypeItemSchemaRemovedUntypedId)
	require.Equal(t, "/drop", removed.GetPath())
	require.Equal(t, checker.ERR, removed.GetLevel())
}

// schema and itemSchema can coexist, so removing the item schema does not
// always leave the response untyped. When a whole-body schema remains the loss
// is smaller and the definition cannot say how much a consumer relied on the
// item shape, so the verdict drops to a warning and carries the reason.
func TestItemSchema_RemovedWhileBodySchemaRemainsIsAWarning(t *testing.T) {
	change := requireChange(t, itemSchemaChanges(t), checker.ResponseBodyMediaTypeItemSchemaRemovedId)

	require.Equal(t, "/keep", change.GetPath())
	require.Equal(t, checker.WARN, change.GetLevel())
	require.Contains(t, change.GetComment(checker.NewDefaultLocalizer()), "still declares a whole-body schema")
}

// The body schema on /drop is unchanged, so nothing may be attributed to it.
func TestItemSchema_DoesNotReportTheBodySchema(t *testing.T) {
	changes := itemSchemaChanges(t)

	require.Nil(t, findChange(changes, checker.RequestBodyMediaTypeSchemaAddedId))
	require.Nil(t, findChange(changes, checker.ResponseBodyMediaTypeSchemaRemovedId))
}
