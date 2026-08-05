package diff

import (
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// OperationsDiff describes the changes between a pair of operation objects (https://swagger.io/specification/#operation-object) of two path item objects
type OperationsDiff struct {
	Added    []string           `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string           `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedOperations `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (operationsDiff *OperationsDiff) Empty() bool {
	if operationsDiff == nil {
		return true
	}

	return len(operationsDiff.Added) == 0 &&
		len(operationsDiff.Deleted) == 0 &&
		len(operationsDiff.Modified) == 0
}

func newOperationsDiff() *OperationsDiff {
	return &OperationsDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedOperations{},
	}
}

// ModifiedOperations is a map of HTTP methods to their respective diffs
type ModifiedOperations map[string]*MethodDiff

func getOperationsDiff(config *Config, state *state, pathItemPair *pathItemPair) (*OperationsDiff, error) {

	if err := filterOperations(config.FilterExtension, pathItemPair); err != nil {
		return nil, err
	}

	diff, err := getOperationsDiffInternal(config, state, pathItemPair)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

// operations are the methods with a dedicated Path Item Object field, in the
// order diff reports them. A document that predates one of them simply has no
// operation under it, so none of them needs a version gate.
var operations = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
	openapi3.MethodQuery, // net/http has no constant for QUERY
}

// methodsToCompare returns the methods to diff for a pair of path items: the
// fixed-field methods above, then any custom method (OpenAPI 3.2
// additionalOperations) either side declares, sorted so output is stable.
//
// The custom methods come from Operations() rather than from
// AdditionalOperations directly, so kin owns which entries count: it drops nil
// entries and lets a fixed field win over a key that shadows it, and a method
// counted twice here would be reported added or deleted twice.
func methodsToCompare(pathItem1, pathItem2 *openapi3.PathItem) []string {
	custom := map[string]struct{}{}
	for _, pathItem := range []*openapi3.PathItem{pathItem1, pathItem2} {
		for method := range pathItem.Operations() {
			if !slices.Contains(operations, method) {
				custom[method] = struct{}{}
			}
		}
	}

	if len(custom) == 0 {
		return operations
	}

	return append(slices.Clone(operations), slices.Sorted(maps.Keys(custom))...)
}

func getOperationsDiffInternal(config *Config, state *state, pathItemPair *pathItemPair) (*OperationsDiff, error) {

	result := newOperationsDiff()
	var err error

	for _, op := range methodsToCompare(pathItemPair.PathItem1, pathItemPair.PathItem2) {
		err = result.diffOperation(config, state, pathItemPair.PathItem1.GetOperation(op), pathItemPair.PathItem2.GetOperation(op), op, pathItemPair.PathParamsMap)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (operationsDiff *OperationsDiff) diffOperation(config *Config, state *state, operation1, operation2 *openapi3.Operation, method string, pathParamsMap PathParamsMap) error {
	if operation1 == nil && operation2 == nil {
		return nil
	}

	if operation1 == nil && operation2 != nil {
		operationsDiff.Added = append(operationsDiff.Added, method)
		return nil
	}

	if operation1 != nil && operation2 == nil {
		operationsDiff.Deleted = append(operationsDiff.Deleted, method)
		return nil
	}

	diff, err := getMethodDiff(config, state, operation1, operation2, pathParamsMap)
	if err != nil {
		return err
	}

	if !diff.Empty() {
		operationsDiff.Modified[method] = diff
	}

	return nil
}

func filterOperations(filterExtension string, pathItemPair *pathItemPair) error {

	if err := filterOperationsByExtensions(filterExtension, pathItemPair); err != nil {
		return err
	}

	return nil
}

func filterOperationsByExtensions(filterExtension string, pathItemPair *pathItemPair) error {
	if filterExtension == "" {
		return nil
	}

	r, err := regexp.Compile(filterExtension)
	if err != nil {
		return fmt.Errorf("failed to compile extension filter regex %q: %w", filterExtension, err)
	}

	filterOperationsByExtensionInternal(pathItemPair.PathItem1, r)
	filterOperationsByExtensionInternal(pathItemPair.PathItem2, r)

	return nil
}

func filterOperationsByExtensionInternal(pathItem *openapi3.PathItem, r *regexp.Regexp) {
	for method, operation := range pathItem.Operations() {
		for extension := range operation.Extensions {
			if r.MatchString(extension) {
				pathItem.SetOperation(method, nil)
				break
			}
		}
	}
}
