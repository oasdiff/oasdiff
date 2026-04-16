package diff

// DependentRequiredDiff describes the changes between a pair of dependentRequired maps.
// Each key in the map is a property name, and the value is the list of properties
// that become required when that property is present.
type DependentRequiredDiff struct {
	Added    map[string][]string `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  map[string][]string `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified map[string]*StringsDiff `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *DependentRequiredDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 && len(diff.Deleted) == 0 && len(diff.Modified) == 0
}

func getDependentRequiredDiff(dr1, dr2 map[string][]string) *DependentRequiredDiff {
	if len(dr1) == 0 && len(dr2) == 0 {
		return nil
	}

	result := &DependentRequiredDiff{
		Added:    map[string][]string{},
		Deleted:  map[string][]string{},
		Modified: map[string]*StringsDiff{},
	}

	for key, values2 := range dr2 {
		values1, ok := dr1[key]
		if !ok {
			result.Added[key] = values2
			continue
		}
		if stringsDiff := getStringsDiff(values1, values2); stringsDiff != nil {
			result.Modified[key] = stringsDiff
		}
	}

	for key, values1 := range dr1 {
		if _, ok := dr2[key]; !ok {
			result.Deleted[key] = values1
		}
	}

	if result.Empty() {
		return nil
	}

	return result
}
