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

// matchedPaths is the result of matching the paths of two specs.
type matchedPaths struct {
	// base and revision are the paths after the prefix rewrites in the config.
	base, revision *openapi3.Paths
	// added, deleted and common are sorted. common holds the base paths of pairs.
	added, deleted, common []string
	pairs                  pathItemPairs
}

// matchPaths applies the prefix rewrites in the config and matches the paths
// of two specs. PathsDiff, EndpointsDiff and the checker contexts all match
// paths through it.
func matchPaths(config *Config, paths1, paths2 *openapi3.Paths) *matchedPaths {

	result := &matchedPaths{
		base:     rewritePrefix(paths1.Map(), config.PathStripPrefixBase, config.PathPrefixBase),
		revision: rewritePrefix(paths2.Map(), config.PathStripPrefixRevision, config.PathPrefixRevision),
		pairs:    pathItemPairs{},
	}

	for _, endpoint1 := range slices.Sorted(maps.Keys(result.base.Map())) {
		if pathItem2, pathParamsMap, ok := findEndpoint(config, endpoint1, result.revision); ok {
			result.common = append(result.common, endpoint1)
			result.pairs[endpoint1] = &pathItemPair{
				PathItem1:     result.base.Value(endpoint1),
				PathItem2:     pathItem2,
				PathParamsMap: pathParamsMap,
			}
		} else {
			result.deleted = append(result.deleted, endpoint1)
		}
	}

	for _, endpoint2 := range slices.Sorted(maps.Keys(result.revision.Map())) {
		if _, _, ok := findEndpoint(config, endpoint2, result.base); !ok {
			result.added = append(result.added, endpoint2)
		}
	}

	return result
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
