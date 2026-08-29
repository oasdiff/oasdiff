package diff

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// Fields of openapi3.Schema that getSchemaDiff deliberately does not compare,
// with the reason. A new kin field is unlisted and uncompared, so it fails
// TestSchemaFieldCoverage until it is either diffed or waived here; that
// turns the next Schema.Always (invisible to the diff for a release, #1190)
// into a build failure instead of a wrong verdict.
var schemaFieldsNotDiffed = map[string]string{
	"Origin": "source-location bookkeeping, not document content",
}

func TestSchemaFieldCoverage(t *testing.T) {
	src := readSchemaDiffSource(t)

	for f := range reflect.TypeFor[openapi3.Schema]().Fields() {
		if !f.IsExported() {
			continue
		}
		if _, waived := schemaFieldsNotDiffed[f.Name]; waived {
			if referenced(src, f.Name) {
				t.Errorf("stale waiver: openapi3.Schema.%s is waived but getSchemaDiff references it; remove the waiver", f.Name)
			}
			continue
		}
		if !referenced(src, f.Name) {
			t.Errorf("openapi3.Schema.%s is not compared by getSchemaDiff\n  diff it or add it to schemaFieldsNotDiffed with a reason", f.Name)
		}
	}
}

// referenced reports whether getSchemaDiff compares the named field: it reads
// value1.X directly, or computes a result field named after it (XDiff), the
// shape used when a helper takes the whole schema, as getRequiredPropertiesDiff
// does for Required.
func referenced(src, field string) bool {
	return strings.Contains(src, "value1."+field) ||
		strings.Contains(src, "result."+field+"Diff =")
}

func readSchemaDiffSource(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("schema_diff.go")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
