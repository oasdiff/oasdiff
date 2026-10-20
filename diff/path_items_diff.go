package diff

import (
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type pathItemPair struct {
	PathItem1     *openapi3.PathItem
	PathItem2     *openapi3.PathItem
	PathParamsMap PathParamsMap
}

type pathItemPairs map[string]*pathItemPair

// getPathItemsDiff matches the paths of two specs. The added and deleted
// paths are sorted, and so are the base paths of the matched pairs.
func getPathItemsDiff(config *Config, paths1, paths2 *openapi3.Paths) (added, deleted, common []string, pairs pathItemPairs) {

	pairs = pathItemPairs{}

	for _, endpoint1 := range slices.Sorted(maps.Keys(paths1.Map())) {
		if pathItem2, pathParamsMap, ok := findEndpoint(config, endpoint1, paths2); ok {
			common = append(common, endpoint1)
			pairs[endpoint1] = &pathItemPair{
				PathItem1:     paths1.Value(endpoint1),
				PathItem2:     pathItem2,
				PathParamsMap: pathParamsMap,
			}
		} else {
			deleted = append(deleted, endpoint1)
		}
	}

	for _, endpoint2 := range slices.Sorted(maps.Keys(paths2.Map())) {
		if _, _, ok := findEndpoint(config, endpoint2, paths1); !ok {
			added = append(added, endpoint2)
		}
	}

	return added, deleted, common, pairs
}

func rewritePrefix(paths map[string]*openapi3.PathItem, strip, prepend string) *openapi3.Paths {
	result := openapi3.NewPathsWithCapacity(len(paths))
	for path, pathItem := range paths {
		result.Set(prepend+strings.TrimPrefix(path, strip), pathItem)
	}
	return result
}

func findEndpoint(config *Config, endpoint string, paths *openapi3.Paths) (*openapi3.PathItem, PathParamsMap, bool) {
	if pathItem := paths.Value(endpoint); pathItem != nil {
		return pathItem, PathParamsMap{}, true
	}

	if config.IncludePathParams {
		return nil, nil, false
	}

	return findNormalizedEndpoint(endpoint, paths)
}

/*
findNormalizedEndpoint finds a corresponding path ignoring differences in template variable names

This implementation is based on Paths.Find in openapi3
*/
func findNormalizedEndpoint(key string, paths *openapi3.Paths) (*openapi3.PathItem, PathParamsMap, bool) {
	normalizedPath, expected, pathParams1 := normalizeTemplatedPath(key)
	pathsMap := paths.Map()
	// several paths can match; visit them in order so the match is the same on every run
	for _, path := range slices.Sorted(maps.Keys(pathsMap)) {
		pathNormalized, got, pathParams2 := normalizeTemplatedPath(path)
		if got == expected && pathNormalized == normalizedPath {
			if pathParamsMap, ok := NewPathParamsMap(pathParams1, pathParams2); ok {
				return pathsMap[path], pathParamsMap, true
			}
		}
	}
	return nil, nil, false
}
