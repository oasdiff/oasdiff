package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

func getTypeDiff(types1, types2 *openapi3.Types) *StringsDiff {
	return getStringsDiff(types1.Slice(), types2.Slice())
}
