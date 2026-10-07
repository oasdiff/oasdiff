package diff

import (
	"cmp"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// SecurityContext carries what a checker needs to work out the security that
// applies to each operation. An operation without its own security list
// inherits the document-root one, so a change to the root list changes
// operations whose own diff is empty, and those operations appear nowhere else
// in the diff.
type SecurityContext struct {
	Base     openapi3.SecurityRequirements
	Revision openapi3.SecurityRequirements

	// Operations holds every operation present in both specs, matched and
	// filtered the same way as PathsDiff, whether or not it changed.
	Operations []OperationPair
}

// OperationPair is an operation present in both specs.
type OperationPair struct {
	Path     string
	Method   string
	Base     *openapi3.Operation
	Revision *openapi3.Operation
}

func (state *state) addOperationPairs(pathItemPairs pathItemPairs) {
	for path, pathItemPair := range pathItemPairs {
		for _, method := range methodsToCompare(pathItemPair.PathItem1, pathItemPair.PathItem2) {
			operation1 := pathItemPair.PathItem1.GetOperation(method)
			operation2 := pathItemPair.PathItem2.GetOperation(method)
			if operation1 == nil || operation2 == nil {
				continue
			}
			state.operationPairs = append(state.operationPairs, OperationPair{
				Path:     path,
				Method:   method,
				Base:     operation1,
				Revision: operation2,
			})
		}
	}
	slices.SortFunc(state.operationPairs, func(a, b OperationPair) int {
		return cmp.Or(cmp.Compare(a.Path, b.Path), cmp.Compare(a.Method, b.Method))
	})
}

// GetSecurityRequirementsDiff calculates the diff between a pair of security
// requirement lists that need not come from the same field, such as the lists
// in effect for an operation before and after it stopped inheriting the
// document-root list.
func GetSecurityRequirementsDiff(securityRequirements1, securityRequirements2 *openapi3.SecurityRequirements) *SecurityRequirementsDiff {
	return getSecurityRequirementsDiff(securityRequirements1, securityRequirements2)
}
