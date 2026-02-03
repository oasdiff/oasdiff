package diff

import (
	"github.com/oasdiff/oasdiff/utils"
)

// InterfaceMap is a map of string to interface
type InterfaceMap map[string]any

// InterfaceMapDiff describes the changes between a pair of InterfaceMap
type InterfaceMapDiff struct {
	Added    utils.StringList   `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList   `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedInterfaces `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *InterfaceMapDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func newInterfaceMapDiff() *InterfaceMapDiff {
	return &InterfaceMapDiff{
		Added:    utils.StringList{},
		Deleted:  utils.StringList{},
		Modified: ModifiedInterfaces{},
	}
}

func getInterfaceMapDiff(map1, map2 InterfaceMap) (*InterfaceMapDiff, error) {
	diff, err := getInterfaceMapDiffInternal(map1, map2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getInterfaceMapDiffInternal(map1, map2 InterfaceMap) (*InterfaceMapDiff, error) {

	result := newInterfaceMapDiff()

	for name1, interface1 := range map1 {
		if interface2, ok := map2[name1]; ok {
			patch, err := compareJson(interface1, interface2)
			if err != nil {
				return nil, err
			}
			if !patch.Empty() {
				result.Modified[name1] = patch
			}
		} else {
			result.Deleted = append(result.Deleted, name1)
		}
	}

	for name2 := range map2 {
		if _, ok := map1[name2]; !ok {
			result.Added = append(result.Added, name2)
		}
	}

	return result, nil
}
