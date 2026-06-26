package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// A single-valued sub-schema (here: an array property's `items`) that exists
// in the base but not the revision produces a sub-schema diff with a nil
// Revision. The property-level checks read the Revision schema (ReadOnly,
// constraints, ...), so walkProperties must not hand them a one-sided diff.
// Regression: this used to panic with a nil pointer dereference, crashing the
// whole process for any spec pair that removed a sub-schema on one side.
func TestWalkPropertiesSkipsOneSidedSubSchemaDiff(t *testing.T) {
	s1, err := open("../data/checker/nil_revision_subschema_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/nil_revision_subschema_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// Every property check runs against a diff whose `items` sub-schema is
	// present only in the base; none of them may panic on the missing side.
	require.NotPanics(t, func() {
		checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)
	})
}

// A media type that existed without a schema and gains one (SchemaDiff.Base nil)
// — or loses its schema (SchemaDiff.Revision nil) — produces a one-sided schema
// diff. The per-schema-change checks read SchemaDiff.Base/Revision, so the
// media-type walkers must not hand them a one-sided diff. Regression for #1047,
// which panicked with a nil pointer dereference when a response media type
// gained a schema. Covers request + response, added + removed.
func TestWalkModifiedSchemasSkipsOneSidedMediaTypeSchema(t *testing.T) {
	base, err := open("../data/media-type-schema/base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/media-type-schema/revision.yaml")
	require.NoError(t, err)

	// Added: request + response media types gain a schema (Base nil).
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	require.NotPanics(t, func() {
		checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)
	})

	// Removed: the same media types lose their schema (Revision nil).
	d, osm, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), revision, base)
	require.NoError(t, err)
	require.NotPanics(t, func() {
		checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)
	})
}
