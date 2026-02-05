package diff

import "github.com/oasdiff/oasdiff/utils"

// StringsDiff describes the changes between a pair of lists of strings
type StringsDiff struct {
	Added   []string `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted []string `json:"deleted,omitempty" yaml:"deleted,omitempty"`
}

func newStringsDiff() *StringsDiff {
	return &StringsDiff{
		Added:   []string{},
		Deleted: []string{},
	}
}

// Empty indicates whether a change was found in this element
func (stringsDiff *StringsDiff) Empty() bool {
	if stringsDiff == nil {
		return true
	}

	return len(stringsDiff.Added) == 0 &&
		len(stringsDiff.Deleted) == 0
}

func getStringsDiff(strings1, strings2 []string) *StringsDiff {
	diff := getStringsDiffInternal(strings1, strings2)

	if diff.Empty() {
		return nil
	}

	return diff
}

func getStringsDiffInternal(strings1, strings2 []string) *StringsDiff {
	result := newStringsDiff()

	s1 := utils.StringSetFromSlice(strings1)
	s2 := utils.StringSetFromSlice(strings2)

	result.Added = s2.Minus(s1).ToStringList()
	result.Deleted = s1.Minus(s2).ToStringList()

	return result
}
