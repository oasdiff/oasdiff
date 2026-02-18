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
