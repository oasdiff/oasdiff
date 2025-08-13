package diff

import (
	"fmt"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

// PathsDiff describes the changes between a pair of Paths objects: https://swagger.io/specification/#paths-object
type PathsDiff struct {
	Added    utils.StringList `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedPaths    `json:"modified,omitempty" yaml:"modified,omitempty"`
	Base     *openapi3.Paths  `json:"-" yaml:"-"`
	Revision *openapi3.Paths  `json:"-" yaml:"-"`
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

	if err := filterPaths(config.MatchPath, config.UnmatchPath, config.FilterExtension, paths1, paths2); err != nil {
		return nil, err
	}

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

	addedPaths, deletedPaths, otherPaths := getPathItemsDiff(config, paths1Mod, paths2Mod)

	for endpoint := range addedPaths.Map() {
		result.addAddedPath(endpoint)
	}

	for endpoint := range deletedPaths.Map() {
		result.addDeletedPath(endpoint)
	}

	for endpoint, pathItemPair := range otherPaths {
		err := result.addModifiedPath(config, state, endpoint, pathItemPair)
		if err != nil {
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

func (pathsDiff *PathsDiff) addAddedPath(path string) {
	pathsDiff.Added = append(pathsDiff.Added, path)
}

func (pathsDiff *PathsDiff) addDeletedPath(path string) {
	pathsDiff.Deleted = append(pathsDiff.Deleted, path)
}

func (pathsDiff *PathsDiff) addModifiedPath(config *Config, state *state, path1 string, pathItemPair *pathItemPair) error {
	return pathsDiff.Modified.addPathDiff(config, state, path1, pathItemPair)
}

func filterPaths(matchPath, unmatchPath, filterExtension string, paths1, paths2 *openapi3.Paths) error {

	if err := filterPathsByName(matchPath, true, paths1, paths2); err != nil {
		return err
	}

	if err := filterPathsByName(unmatchPath, false, paths1, paths2); err != nil {
		return err
	}

	if err := filterPathsByExtensions(filterExtension, paths1, paths2); err != nil {
		return err
	}

	return nil
}

func filterPathsByName(filter string, negate bool, paths1, paths2 *openapi3.Paths) error {
	if filter == "" {
		return nil
	}

	r, err := regexp.Compile(filter)
	if err != nil {
		return fmt.Errorf("failed to compile filter regex %q: %w", filter, err)
	}

	filterPathsInternal(paths1, r, negate)
	filterPathsInternal(paths2, r, negate)

	return nil
}

func filterPathsInternal(paths *openapi3.Paths, r *regexp.Regexp, negate bool) {
	for path := range paths.Map() {
		match := r.MatchString(path)
		if negate {
			match = !match
		}
		if match {
			paths.Delete(path)
		}
	}
}

func filterPathsByExtensions(filterExtension string, paths1, paths2 *openapi3.Paths) error {
	if filterExtension == "" {
		return nil
	}

	r, err := regexp.Compile(filterExtension)
	if err != nil {
		return fmt.Errorf("failed to compile extension filter regex %q: %w", filterExtension, err)
	}

	filterPathsByExtensionInternal(paths1, r)
	filterPathsByExtensionInternal(paths2, r)

	return nil
}

func filterPathsByExtensionInternal(paths *openapi3.Paths, r *regexp.Regexp) {
	for path, pathItem := range paths.Map() {
		for extension := range pathItem.Extensions {
			if r.MatchString(extension) {
				paths.Delete(path)
				break
			}
		}
	}
}
