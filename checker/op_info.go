package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// opInfo bundles the per-operation plumbing that every helper needs
// (config, operation, operationsSources, method, path) so signatures
// stay focused on what actually varies. Used as an argument in helper
// functions in order to simplify the function signature.
//
// methodDiff is set only when the helper is operating on a diff
// (base vs revision MethodDiff). Single-side checks like sunset leave
// it nil and use `operation` directly.
type opInfo struct {
	config            *Config
	operation         *openapi3.Operation
	operationsSources *diff.OperationsSourcesMap
	method            string
	path              string
	// methodDiff is the diff record for this operation (base vs
	// revision). Set by newOpInfoFromDiff; nil when the helper is
	// operating on a single openapi3.Operation rather than a diff.
	methodDiff *diff.MethodDiff
}

func newOpInfo(config *Config, operation *openapi3.Operation, operationsSources *diff.OperationsSourcesMap, method, path string) opInfo {
	return opInfo{
		config:            config,
		operation:         operation,
		operationsSources: operationsSources,
		method:            method,
		path:              path,
	}
}

// newOpInfoFromDiff is the diff-helper constructor: takes the
// MethodDiff and uses its Revision side as the underlying
// *openapi3.Operation (matching the existing convention used by the
// list-of-types and enum-diff helpers, which always pass
// operationItem.Revision to NewApiChange).
func newOpInfoFromDiff(config *Config, methodDiff *diff.MethodDiff, operationsSources *diff.OperationsSourcesMap, method, path string) opInfo {
	return opInfo{
		config:            config,
		operation:         methodDiff.Revision,
		operationsSources: operationsSources,
		method:            method,
		path:              path,
		methodDiff:        methodDiff,
	}
}
