package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// itemSchema (OpenAPI 3.2) describes one item of a streamed body. It is diffed
// with the same schema-diff engine as schema, so it inherits property, type,
// constraint and enum comparison rather than reimplementing any of it.
func TestDiff_ItemSchema(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/item-schema/base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/item-schema/revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	requestItem := d.PathsDiff.Modified["/ingest"].OperationsDiff.Modified["POST"].
		RequestBodyDiff.ContentDiff.MediaTypeModified["application/jsonl"].ItemSchemaDiff
	require.NotNil(t, requestItem)
	require.Contains(t, requestItem.RequiredDiff.Added, "weight")

	responseItem := d.PathsDiff.Modified["/events"].OperationsDiff.Modified["GET"].
		ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["text/event-stream"].ItemSchemaDiff
	require.NotNil(t, responseItem)
	require.NotEmpty(t, responseItem.PropertiesDiff.Modified)

	// An itemSchema present on one side only is flagged, not diffed.
	dropped := d.PathsDiff.Modified["/drop"].OperationsDiff.Modified["POST"].
		ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/jsonl"].ItemSchemaDiff
	require.True(t, dropped.SchemaDeleted)
}
