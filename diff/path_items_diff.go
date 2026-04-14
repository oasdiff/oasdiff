package diff

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type pathItemPair struct {
	PathItem1     *openapi3.PathItem
	PathItem2     *openapi3.PathItem
	PathParamsMap PathParamsMap
}

type pathItemPairs map[string]*pathItemPair

func getPathItemsDiff(config *Config, paths1, paths2 *openapi3.Paths) (openapi3.Paths, openapi3.Paths, pathItemPairs) {

	added := openapi3.Paths{}
	deleted := openapi3.Paths{}
	other := pathItemPairs{}

	for endpoint1, pathItem1 := range paths1.Map() {
		if pathItem2, pathParamsMap, ok := findEndpoint(config, endpoint1, paths2); ok {
			other[endpoint1] = &pathItemPair{
				PathItem1:     pathItem1,
				PathItem2:     pathItem2,
				PathParamsMap: pathParamsMap,
			}
		} else {
			deleted.Set(endpoint1, pathItem1)
		}
	}

	for endpoint2, pathItem2 := range paths2.Map() {
		if _, _, ok := findEndpoint(config, endpoint2, paths1); !ok {
			added.Set(endpoint2, pathItem2)
		}
	}

	return added, deleted, other
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
	for path, pathItem := range paths.Map() {
		pathNormalized, got, pathParams2 := normalizeTemplatedPath(path)
		if got == expected && pathNormalized == normalizedPath {
			if pathParamsMap, ok := NewPathParamsMap(pathParams1, pathParams2); ok {
				return pathItem, pathParamsMap, true
			}
		}
	}
	return nil, nil, false
}
