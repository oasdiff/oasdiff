package diff

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

// OperationsDiff describes the changes between a pair of operation objects (https://swagger.io/specification/#operation-object) of two path item objects
type OperationsDiff struct {
	Added    utils.StringList   `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList   `json:"deleted,omitempty" yaml:"deleted,omitempty"`
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
		Added:    utils.StringList{},
		Deleted:  utils.StringList{},
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
}

func getOperationsDiffInternal(config *Config, state *state, pathItemPair *pathItemPair) (*OperationsDiff, error) {

	result := newOperationsDiff()
	var err error

	for _, op := range operations {
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
