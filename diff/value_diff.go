package diff

import (
	"reflect"
)

// ValueDiff describes the changes between a pair of values
type ValueDiff struct {
	From any `json:"from" yaml:"from"`
	To   any `json:"to" yaml:"to"`
}

// Empty indicates whether a change was found in this element
func (diff *ValueDiff) Empty() bool {
	return diff == nil
}

func getValueDiff(value1, value2 any) *ValueDiff {

	diff := getValueDiffInternal(value1, value2)

	if diff.Empty() {
		return nil
	}

	return diff
}

func getValueDiffInternal(value1, value2 any) *ValueDiff {
	if reflect.DeepEqual(value1, value2) {
		return nil
	}

	return &ValueDiff{
		From: value1,
		To:   value2,
	}
}

func getValueDiffConditional(exclude bool, value1, value2 any) *ValueDiff {
	if exclude {
		return nil
	}

	return getValueDiff(value1, value2)
}

func getFloat64RefDiff(valueRef1, valueRef2 *float64) *ValueDiff {
	return getValueDiff(derefFloat64(valueRef1), derefFloat64(valueRef2))
}

func getBoolRefDiff(valueRef1, valueRef2 *bool) *ValueDiff {
	return getValueDiff(derefBool(valueRef1), derefBool(valueRef2))
}

func getStringRefDiffConditional(exclude bool, valueRef1, valueRef2 *string) *ValueDiff {
	return getValueDiffConditional(exclude, derefString(valueRef1), derefString(valueRef2))
}

func getUInt64RefDiff(valueRef1, valueRef2 *uint64) *ValueDiff {
	return getValueDiff(derefUInt64(valueRef1), derefUInt64(valueRef2))
}

func derefString(ref *string) any {
	if ref == nil {
		return nil
	}

	return *ref
}

func derefBool(ref *bool) any {
	if ref == nil {
		return nil
	}

	return *ref
}

func derefFloat64(ref *float64) any {
	if ref == nil {
		return nil
	}

	return *ref
}

func derefUInt64(ref *uint64) any {
	if ref == nil {
		return nil
	}

	return *ref
}

// exclusiveBoundToValue extracts the meaningful value from an ExclusiveBound.
// Returns the numeric value if set (OpenAPI 3.1 style), or the boolean if set (OpenAPI 3.0 style).
func exclusiveBoundToValue(eb any) any {
	// Use type assertion to handle the struct
	// We use 'any' to avoid import cycle with openapi3
	type exclusiveBound interface {
		IsSet() bool
	}

	if eb == nil {
		return nil
	}

	// Check if it implements our interface
	if checker, ok := eb.(exclusiveBound); ok && !checker.IsSet() {
		return nil
	}

	return eb
}

// getExclusiveBoundDiff compares two ExclusiveBound values
func getExclusiveBoundDiff(eb1, eb2 any) *ValueDiff {
	v1 := exclusiveBoundToValue(eb1)
	v2 := exclusiveBoundToValue(eb2)
	return getValueDiff(v1, v2)
}
