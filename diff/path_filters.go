package diff

import (
	"fmt"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
)

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

func filterOperationsByExtensions(r *regexp.Regexp, pathItemPair *pathItemPair) {
	if r == nil {
		return
	}

	filterOperationsByExtensionInternal(pathItemPair.PathItem1, r)
	filterOperationsByExtensionInternal(pathItemPair.PathItem2, r)
}

func filterOperationsByExtensionInternal(pathItem *openapi3.PathItem, r *regexp.Regexp) {
	for method, operation := range pathItem.Operations() {
		for extension := range operation.Extensions {
			if r.MatchString(extension) {
				pathItem.SetOperation(method, nil)
				break
			}
		}
	}
}
