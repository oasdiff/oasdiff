package internal

import (
	"fmt"
	"strings"

	"slices"
)

type enumVal interface {
	Set(s string) error
	String() string
	Type() string
	listOf() string
}

// enumValue is like stringValue with allowed values
type enumValue struct {
	value         *string
	allowedValues []string
}

func newEnumValue(allowedValues []string, val string) *enumValue {
	result := new(enumValue)
	result.allowedValues = allowedValues
	result.value = &val
	return result
}

// String is used both by fmt.Print and by Cobra in help text
func (v *enumValue) String() string {
	return string(*v.value)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (v *enumValue) Set(s string) error {
	if slices.Contains(v.allowedValues, s) {
		*v.value = s
		return nil
	}
	return fmt.Errorf("%s is not one of the allowed values: %s", s, v.listOf())
}

func (v *enumValue) listOf() string {
	l := len(v.allowedValues)
	switch l {
	case 0:
		return "no options available"
	case 1:
		return v.allowedValues[0]
	case 2:
		return v.allowedValues[0] + " or " + v.allowedValues[1]
	default:
		return strings.Join(v.allowedValues[:l-1], ", ") + ", or " + v.allowedValues[l-1]
	}
}

// Type is only used in help text
func (v *enumValue) Type() string {
	return "string"
}
