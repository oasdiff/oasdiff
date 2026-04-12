package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// opInfo is used as an argument in helper functions in order to simplify the function signature.
type opInfo struct {
	config            *Config
	operation         *openapi3.Operation
	operationsSources *diff.OperationsSourcesMap
	method            string
	path              string
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
