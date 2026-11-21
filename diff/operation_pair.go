package diff

import (
	"cmp"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// OperationPair is an operation present in both specs.
type OperationPair struct {
	// Webhook is true for an operation of a webhook. Path is then the
	// webhook's name.
	Webhook  bool
	Path     string
	Method   string
	Base     *openapi3.Operation
	Revision *openapi3.Operation
}

// getOperationPairs returns the operations of pathItemPairs, sorted by path, or
// webhook name, and method.
func getOperationPairs(pathItemPairs pathItemPairs, webhook bool) []OperationPair {
	result := []OperationPair{}
	for path, pathItemPair := range pathItemPairs {
		for _, method := range methodsToCompare(pathItemPair.PathItem1, pathItemPair.PathItem2) {
			operation1 := pathItemPair.PathItem1.GetOperation(method)
			operation2 := pathItemPair.PathItem2.GetOperation(method)
			if operation1 == nil || operation2 == nil {
				continue
			}
			result = append(result, OperationPair{
				Webhook:  webhook,
				Path:     path,
				Method:   method,
				Base:     operation1,
				Revision: operation2,
			})
		}
	}
	slices.SortFunc(result, func(a, b OperationPair) int {
		return cmp.Or(cmp.Compare(a.Path, b.Path), cmp.Compare(a.Method, b.Method))
	})
	return result
}
