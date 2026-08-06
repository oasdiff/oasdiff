package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// The operations list decides both what gets diffed and the order it is
// reported in, so it has to stay a superset of the methods kin gives a
// dedicated Path Item field.
//
// A method kin promotes to a fixed field and this list has not caught up with
// is still diffed, since methodsToCompare falls back to treating it as custom,
// but it lands in the sorted tail rather than in a position chosen here. This
// fails when that happens so the position is picked deliberately.
func TestOperations_CoverKinFixedFields(t *testing.T) {
	for _, method := range openapi3.PathItemMethods() {
		require.Contains(t, operations, method,
			"kin has a dedicated Path Item field for %s; add it to `operations` in the position diff should report it", method)
	}
}
