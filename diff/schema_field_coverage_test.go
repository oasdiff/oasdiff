package diff

import (
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/internal/populatetest"
	"github.com/stretchr/testify/require"
)

// Fields of openapi3.Schema that getSchemaDiff deliberately does not compare,
// with the reason.
var schemaFieldsNotDiffed = map[string]string{
	"Origin": "source-location bookkeeping, not document content",
}

// TestSchemaFieldCoverage sets each exported field of openapi3.Schema, one at
// a time, on one side of an otherwise-empty schema pair and requires the diff
// to see the change. A field kin adds is set like any other, so it fails here
// until it is compared or waived above — a wrong verdict (Schema.Always was
// invisible to the diff for a release, #1190) becomes a build failure
// instead. Same gate as flatten/allof's TestMerge_SingleSchema_PreservesEveryField,
// asked of the diff.
func TestSchemaFieldCoverage(t *testing.T) {
	cfg := NewConfig()
	for f := range reflect.TypeFor[openapi3.Schema]().Fields() {
		if !f.IsExported() {
			continue
		}
		if _, waived := schemaFieldsNotDiffed[f.Name]; waived {
			continue
		}

		revision := &openapi3.Schema{}
		if !populatetest.NonZero(reflect.ValueOf(revision).Elem().FieldByName(f.Name), f.Name) {
			// Silently skipping would reopen the hole this test exists to
			// close: a field of a kind setNonZeroValue cannot build would go
			// unchecked and the test would still pass.
			t.Errorf("cannot populate %s (%s); extend populatetest.NonZero so the field is covered", f.Name, f.Type)
			continue
		}

		d, err := getSchemaDiff(cfg, newState(),
			&openapi3.SchemaRef{Value: &openapi3.Schema{}},
			&openapi3.SchemaRef{Value: revision})
		require.NoError(t, err)
		if d.Empty() {
			t.Errorf("openapi3.Schema.%s changed but the diff is empty\n  compare it in getSchemaDiff or add it to schemaFieldsNotDiffed with a reason", f.Name)
		}
	}
}
