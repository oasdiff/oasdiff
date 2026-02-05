package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// CallbacksDiff describes the changes between a pair of callback objects: https://swagger.io/specification/#callback-object
type CallbacksDiff struct {
	Added    []string  `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string  `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedCallbacks `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *CallbacksDiff) Empty() bool {

	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

// ModifiedCallbacks is map of callback names to their respective diffs
type ModifiedCallbacks map[string]*PathsDiff

func newCallbacksDiff() *CallbacksDiff {
	return &CallbacksDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedCallbacks{},
	}
}

func getCallbacksDiff(config *Config, state *state, callbacks1, callbacks2 openapi3.Callbacks) (*CallbacksDiff, error) {
	diff, err := getCallbacksDiffInternal(config, state, callbacks1, callbacks2)

	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getCallbacksDiffInternal(config *Config, state *state, callbacks1, callbacks2 openapi3.Callbacks) (*CallbacksDiff, error) {

	result := newCallbacksDiff()

	for callbackName1, callbackRef1 := range callbacks1 {
		if callbackRef2, ok := callbacks2[callbackName1]; ok {

			value1, err := derefCallback(callbackRef1)
			if err != nil {
				return nil, err
			}
			value2, err := derefCallback(callbackRef2)
			if err != nil {
				return nil, err
			}

			diff, err := getCallbackDiff(config, state, value1, value2)
			if err != nil {
				return nil, err
			}

			if !diff.Empty() {
				result.Modified[callbackName1] = diff
			}
		} else {
			result.Deleted = append(result.Deleted, callbackName1)
		}
	}

	for callbackValue2 := range callbacks2 {
		if _, ok := callbacks1[callbackValue2]; !ok {
			result.Added = append(result.Added, callbackValue2)
		}
	}

	return result, nil

}

func derefCallback(ref *openapi3.CallbackRef) (*openapi3.Callback, error) {

	if ref == nil || ref.Value == nil {
		return nil, fmt.Errorf("callback reference is nil")
	}

	return ref.Value, nil
}

func getCallbackDiff(config *Config, state *state, callback1, callback2 *openapi3.Callback) (*PathsDiff, error) {
	return getPathsDiff(config, state, callBackToPaths(callback1), callBackToPaths(callback2))
}

func callBackToPaths(callback *openapi3.Callback) *openapi3.Paths {
	result := openapi3.NewPathsWithCapacity(callback.Len())
	for k, v := range callback.Map() {
		result.Set(k, v)
	}
	return result
}

func (diff *CallbacksDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
