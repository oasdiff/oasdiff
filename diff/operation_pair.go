package diff

import (
	"cmp"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// OperationPair is an operation present in both specs.
type OperationPair struct {
	Path     string
	Method   string
	Base     *openapi3.Operation
	Revision *openapi3.Operation
}

func getOperationPairs(pathItemPairs pathItemPairs) []OperationPair {
	result := []OperationPair{}
	for path, pathItemPair := range pathItemPairs {
		for _, method := range methodsToCompare(pathItemPair.PathItem1, pathItemPair.PathItem2) {
			operation1 := pathItemPair.PathItem1.GetOperation(method)
			operation2 := pathItemPair.PathItem2.GetOperation(method)
			if operation1 == nil || operation2 == nil {
				continue
			}
			result = append(result, OperationPair{
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
