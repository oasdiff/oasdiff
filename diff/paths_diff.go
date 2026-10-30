package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// PathsDiff describes the changes between a pair of Paths objects: https://swagger.io/specification/#paths-object
type PathsDiff struct {
	Added    []string        `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string        `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedPaths   `json:"modified,omitempty" yaml:"modified,omitempty"`
	Base     *openapi3.Paths `json:"-" yaml:"-"`
	Revision *openapi3.Paths `json:"-" yaml:"-"`
}

// Empty indicates whether a change was found in this element
func (pathsDiff *PathsDiff) Empty() bool {
	if pathsDiff == nil {
		return true
	}

	return len(pathsDiff.Added) == 0 &&
		len(pathsDiff.Deleted) == 0 &&
		len(pathsDiff.Modified) == 0
}

func newPathsDiff() *PathsDiff {
	return &PathsDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedPaths{},
	}
}

func getPathsDiff(config *Config, state *state, paths1, paths2 *openapi3.Paths) (*PathsDiff, error) {

	filters, err := getFilters(config, state)
	if err != nil {
		return nil, err
	}
	filterPaths(filters, paths1, paths2)

	diff, err := getPathsDiffInternal(config, state, paths1, paths2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getPathsDiffInternal(config *Config, state *state, paths1, paths2 *openapi3.Paths) (*PathsDiff, error) {

	result := newPathsDiff()

	paths1Mod := rewritePrefix(paths1.Map(), config.PathStripPrefixBase, config.PathPrefixBase)
	paths2Mod := rewritePrefix(paths2.Map(), config.PathStripPrefixRevision, config.PathPrefixRevision)

	added, deleted, common, pairs := getPathItemsDiff(config, paths1Mod, paths2Mod)
	result.Added = added
	result.Deleted = deleted

	for _, endpoint := range common {
		if err := result.addModifiedPath(config, state, endpoint, pairs[endpoint]); err != nil {
			return nil, err
		}
	}
	result.Base = paths1Mod
	result.Revision = paths2Mod

	return result, nil
}

func (pathsDiff *PathsDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(pathsDiff.Added),
		Deleted:  len(pathsDiff.Deleted),
		Modified: len(pathsDiff.Modified),
	}
}

func (pathsDiff *PathsDiff) addModifiedPath(config *Config, state *state, path1 string, pathItemPair *pathItemPair) error {
	return pathsDiff.Modified.addPathDiff(config, state, path1, pathItemPair)
}
