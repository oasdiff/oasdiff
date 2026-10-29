package diff

import (
	"fmt"
	"regexp"

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

// pathFilters are the compiled --match-path, --unmatch-path and
// --filter-extension patterns. A nil pattern means that filter is off.
type pathFilters struct {
	matchPath   *regexp.Regexp
	unmatchPath *regexp.Regexp
	extension   *regexp.Regexp
}

// getFilters compiles the config's filter patterns and keeps them on the
// state, so each pattern is compiled once per diff. The operations filter is
// applied to every path pair being compared.
func getFilters(config *Config, state *state) (*pathFilters, error) {
	if state.filters != nil {
		return state.filters, nil
	}

	filters := &pathFilters{}

	var err error
	if filters.matchPath, err = compileFilter(config.MatchPath, "filter"); err != nil {
		return nil, err
	}
	if filters.unmatchPath, err = compileFilter(config.UnmatchPath, "filter"); err != nil {
		return nil, err
	}
	if filters.extension, err = compileFilter(config.FilterExtension, "extension filter"); err != nil {
		return nil, err
	}

	state.filters = filters

	return filters, nil
}

// compileFilter compiles one filter pattern. An empty pattern compiles to nil,
// which leaves that filter off.
func compileFilter(pattern, name string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}

	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile %s regex %q: %w", name, pattern, err)
	}

	return r, nil
}

func filterPaths(filters *pathFilters, paths1, paths2 *openapi3.Paths) {

	filterPathsByName(filters.matchPath, true, paths1, paths2)
	filterPathsByName(filters.unmatchPath, false, paths1, paths2)
	filterPathsByExtensions(filters.extension, paths1, paths2)
}

func filterPathsByName(r *regexp.Regexp, negate bool, paths1, paths2 *openapi3.Paths) {
	if r == nil {
		return
	}

	filterPathsInternal(paths1, r, negate)
	filterPathsInternal(paths2, r, negate)
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

func filterPathsByExtensions(r *regexp.Regexp, paths1, paths2 *openapi3.Paths) {
	if r == nil {
		return
	}

	filterPathsByExtensionInternal(paths1, r)
	filterPathsByExtensionInternal(paths2, r)
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
