package diff

// InterfaceMap is a map of string to interface
type InterfaceMap map[string]any

// InterfaceMapDiff describes the changes between a pair of InterfaceMap
type InterfaceMapDiff struct {
	Added    []string           `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string           `json:"deleted,omitempty" yaml:"deleted,omitempty"`
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
		Added:    []string{},
		Deleted:  []string{},
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

	added, deleted, common := keysDiff(map1, map2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		patch, err := compareJson(map1[name], map2[name])
		if err != nil {
			return nil, err
		}
		if !patch.Empty() {
			result.Modified[name] = patch
		}
	}

	return result, nil
}
