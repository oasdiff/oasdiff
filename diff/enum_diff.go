package diff

import (
	"reflect"
)

// EnumDiff describes the changes between a pair of enums
type EnumDiff struct {
	EnumAdded   bool       `json:"enumAdded,omitempty" yaml:"enumAdded,omitempty"`
	EnumDeleted bool       `json:"enumDeleted,omitempty" yaml:"enumDeleted,omitempty"`
	Added       EnumValues `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted     EnumValues `json:"deleted,omitempty" yaml:"deleted,omitempty"`
}

// EnumValues is a list of enum values
type EnumValues []interface{}

func newEnumDiff() *EnumDiff {
	return &EnumDiff{
		Added:   EnumValues{},
		Deleted: EnumValues{},
	}
}

// Empty indicates whether a change was found in this element
func (enumDiff *EnumDiff) Empty() bool {
	if enumDiff == nil {
		return true
	}

	return len(enumDiff.Added) == 0 &&
		len(enumDiff.Deleted) == 0
}

func getEnumDiff(enum1, enum2 EnumValues) *EnumDiff {

	diff := getEnumDiffInternal(enum1, enum2)

	if diff.Empty() {
		return nil
	}

	return diff
}

func getEnumDiffInternal(enum1, enum2 EnumValues) *EnumDiff {

	diff := newEnumDiff()

	if enum1 == nil && enum2 != nil {
		diff.EnumAdded = true
	}

	if enum1 != nil && enum2 == nil {
		diff.EnumDeleted = true
	}

	for _, v1 := range enum1 {
		if !findValue(v1, enum2) {
			diff.Deleted = append(diff.Deleted, v1)
		}
	}

	for _, v2 := range enum2 {
		if !findValue(v2, enum1) {
			diff.Added = append(diff.Added, v2)
		}
	}

	return diff
}

func findValue(value interface{}, enum EnumValues) bool {
	for _, other := range enum {
		if reflect.DeepEqual(value, other) {
			return true
		}
	}
	return false
}

// Patch applies the patch to an enum
func (enumDiff *EnumDiff) Patch(enum *[]interface{}) {

	if enumDiff.Empty() {
		return
	}

	result := []interface{}{}

	for _, value := range *enum {
		if !findValue(value, enumDiff.Deleted) {
			result = append(result, value)
		}
	}

	for _, value := range enumDiff.Added {
		result = append(result, value)
	}

	*enum = result
}
