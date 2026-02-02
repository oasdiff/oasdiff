package utils

import (
	"maps"
	"slices"
)

// StringSet is a set of string values
type StringSet map[string]struct{}

func (stringSet StringSet) ToStringList() StringList {
	return StringList(slices.Collect(maps.Keys(stringSet)))
}

func (stringSet StringSet) Add(s string) {
	stringSet[s] = struct{}{}
}

func (stringSet StringSet) Contains(s string) bool {
	_, ok := stringSet[s]
	return ok
}

func (stringSet StringSet) Minus(other StringSet) StringSet {
	result := StringSet{}

	for s := range stringSet {
		if !other.Contains(s) {
			result.Add(s)
		}
	}

	return result
}

func (stringSet StringSet) Plus(other StringSet) StringSet {
	result := stringSet.Copy()

	for s := range other {
		result.Add(s)
	}

	return result
}

func (stringSet StringSet) Intersection(other StringSet) StringSet {
	result := StringSet{}

	for s := range stringSet {
		if other.Contains(s) {
			result.Add(s)
		}
	}

	return result
}

func (stringSet StringSet) Equals(other StringSet) bool {
	return stringSet.Minus(other).Empty() &&
		other.Minus(stringSet).Empty()
}

// Empty indicates whether a change was found in this element
func (stringSet StringSet) Empty() bool {
	return len(stringSet) == 0
}

func (stringSet StringSet) Copy() StringSet {
	result := make(StringSet, len(stringSet))
	for k := range stringSet {
		result.Add(k)
	}
	return result
}
