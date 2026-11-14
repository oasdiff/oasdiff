package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
)

// OperationsSourcesMap maps OpenAPI operations to the spec each was loaded
// from. The value is the source itself, not its name, so a consumer can tell
// a file from a url, a git revision or stdin.
type OperationsSourcesMap map[*openapi3.Operation]*load.Source

// Source returns the spec an operation was loaded from, and nil for an
// operation the map does not have. The load.Source accessors tolerate a nil
// receiver, so a caller can ask a missing operation what kind of source it
// came from and get the answers for none.
func (operationsSources OperationsSourcesMap) Source(operation *openapi3.Operation) *load.Source {
	return operationsSources[operation]
}
