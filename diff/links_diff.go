package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// LinksDiff describes the changes between a pair of sets of link objects: https://swagger.io/specification/#link-object
type LinksDiff struct {
	Added    []string      `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string      `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedLinks `json:"modified,omitempty" yaml:"modified,omitempty"`
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
		Added:    []string{},
		Deleted:  []string{},
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

	added, deleted, common := keysDiff(links1, links2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		value1, err := derefLink(links1[name])
		if err != nil {
			return nil, err
		}

		value2, err := derefLink(links2[name])
		if err != nil {
			return nil, err
		}

		diff, err := getLinkDiff(config, value1, value2)
		if err != nil {
			return nil, err
		}
		if !diff.Empty() {
			result.Modified[name] = diff
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
