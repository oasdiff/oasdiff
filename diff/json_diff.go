package diff

import (
	"fmt"

	"github.com/wI2L/jsondiff"
)

// JsonPatch is a wrapper to jsondiff.JsonPatch with proper serialization for json and yaml
type JsonPatch []*JsonOperation

// Empty indicates whether a change was found in this element
func (p JsonPatch) Empty() bool {
	return len(p) == 0
}

// JsonOperation is a wrapper to jsondiff.JsonOperation with proper serialization for json and yaml
type JsonOperation struct {
	OldValue any    `json:"oldValue" yaml:"oldValue"`
	Value    any    `json:"value" yaml:"value"`
	Type     string `json:"op" yaml:"op"`
	From     string `json:"from" yaml:"from"`
	Path     string `json:"path" yaml:"path"`
}

func (op *JsonOperation) String() string {
	switch op.Type {
	case "add":
		return fmt.Sprintf("Added %s with value: '%v'", op.Path, op.Value)
	case "remove":
		return fmt.Sprintf("Removed %s with value: '%v'", op.Path, op.OldValue)
	case "replace":
		path := op.Path
		if path == "" {
			path = "value"
		}
		return fmt.Sprintf("Modified %s from '%v' to '%v'", path, op.OldValue, op.Value)
	default:
		return fmt.Sprintf("%s %s", op.Type, op.Path)
	}
}

func toJsonPatch(patch jsondiff.Patch) JsonPatch {
	result := make(JsonPatch, len(patch))
	for i, op := range patch {
		result[i] = newJsonOperation(op)
	}
	return result
}

func newJsonOperation(op jsondiff.Operation) *JsonOperation {
	return &JsonOperation{
		OldValue: op.OldValue,
		Value:    op.Value,
		Type:     op.Type,
		From:     op.From,
		Path:     op.Path,
	}
}

func compareJson(source, target any, opts ...jsondiff.Option) (JsonPatch, error) {
	patch, err := jsondiff.Compare(source, target, opts...)
	if err != nil {
		return nil, err
	}
	return toJsonPatch(patch), nil
}
