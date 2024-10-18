package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/utils"
)

// LinksDiff describes the changes between a pair of sets of link objects: https://swagger.io/specification/#link-object
type LinksDiff struct {
	Added    utils.StringList `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedLinks    `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *LinksDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

// ModifiedLinks is map of link values to their respective diffs
type ModifiedLinks map[string]*LinkDiff

func newLinksDiff() *LinksDiff {
	return &LinksDiff{
		Added:    utils.StringList{},
		Deleted:  utils.StringList{},
		Modified: ModifiedLinks{},
	}
}

func getLinksDiff(config *Config, links1, links2 openapi3.Links) (*LinksDiff, error) {
	diff, err := getLinksDiffsInternal(config, links1, links2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getLinksDiffsInternal(config *Config, links1, links2 openapi3.Links) (*LinksDiff, error) {

	result := newLinksDiff()

	for name1, ref1 := range links1 {
		if ref2, ok := links2[name1]; ok {
			value1, err := derefLink(ref1)
			if err != nil {
				return nil, err
			}

			value2, err := derefLink(ref2)
			if err != nil {
				return nil, err
			}

			diff, err := getLinkDiff(config, value1, value2)
			if err != nil {
				return nil, err
			}
			if !diff.Empty() {
				result.Modified[name1] = diff
			}
		} else {
			result.Deleted = append(result.Deleted, name1)
		}
	}

	for name2 := range links2 {
		if _, ok := links1[name2]; !ok {
			result.Added = append(result.Added, name2)
		}
	}

	return result, nil
}

func derefLink(ref *openapi3.LinkRef) (*openapi3.Link, error) {

	if ref == nil || ref.Value == nil {
		return nil, fmt.Errorf("link reference is nil")
	}

	return ref.Value, nil
}

func (diff *LinksDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
