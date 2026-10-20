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

func getOperationPairs(config *Config, paths1, paths2 *openapi3.Paths) []OperationPair {
	paths1Mod := rewritePrefix(paths1.Map(), config.PathStripPrefixBase, config.PathPrefixBase)
	paths2Mod := rewritePrefix(paths2.Map(), config.PathStripPrefixRevision, config.PathPrefixRevision)
	_, _, _, pathItemPairs := getPathItemsDiff(config, paths1Mod, paths2Mod)

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
