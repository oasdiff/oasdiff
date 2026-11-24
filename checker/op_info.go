package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

// opInfo bundles the per-operation plumbing that every helper needs
// (config, operation, operationsSources, method, path) so signatures
// stay focused on what actually varies. Used as an argument in helper
// functions in order to simplify the function signature.
//
// operations is set only when the helper is comparing the two sides of
// an operation. Single-side checks like sunset leave it nil and use
// `operation` directly.
type opInfo struct {
	config            *Config
	operation         *openapi3.Operation
	operationsSources diff.OperationsSourcesMap
	method            string
	path              string
	operations        location.Operations
}

// NewApiChange builds an ApiChange from the operation context carried by o,
// so a caller supplies only what varies per change (id, args, comment) instead
// of repeating config/operationsSources/operation/method/path at every call.
// The caller still chains WithSources / WithDetails as needed.
func (o opInfo) NewApiChange(id string, args []any, comment string) ApiChange {
	return NewApiChange(id, o.config, args, comment, o.operationsSources, o.operation, o.method, o.path)
}

func newOpInfo(config *Config, operation *openapi3.Operation, operationsSources diff.OperationsSourcesMap, method, path string) opInfo {
	return opInfo{
		config:            config,
		operation:         operation,
		operationsSources: operationsSources,
		method:            method,
		path:              path,
	}
}

// newOpInfoFromDiff is the constructor for a helper that compares the two
// sides of an operation: a *diff.MethodDiff, or a diff.OperationPair. The
// revision side is the operation changes are reported on.
func newOpInfoFromDiff(config *Config, operations location.Operations, operationsSources diff.OperationsSourcesMap, method, path string) opInfo {
	return opInfo{
		config:            config,
		operation:         operations.RevisionOperation(),
		operationsSources: operationsSources,
		method:            method,
		path:              path,
		operations:        operations,
	}
}
