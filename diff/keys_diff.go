package diff

import (
	"maps"
	"slices"
)

// keysDiff compares the keys of two maps. Each list is sorted, since Go
// randomizes map iteration and the diff must come out the same on every run.
func keysDiff[V any](m1, m2 map[string]V) (added, deleted, common []string) {
	for _, k := range slices.Sorted(maps.Keys(m1)) {
		if _, ok := m2[k]; ok {
			common = append(common, k)
		} else {
			deleted = append(deleted, k)
		}
	}
	for _, k := range slices.Sorted(maps.Keys(m2)) {
		if _, ok := m1[k]; !ok {
			added = append(added, k)
		}
	}
	return added, deleted, common
}
