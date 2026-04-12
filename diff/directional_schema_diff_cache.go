package diff

import "github.com/getkin/kin-openapi/openapi3"

type directionalSchemaDiffCache struct {
	requestCache  schemaDiffCache
	responseCache schemaDiffCache
}

func newDirectionalSchemaDiffCache() directionalSchemaDiffCache {
	return directionalSchemaDiffCache{
		requestCache:  schemaDiffCache{},
		responseCache: schemaDiffCache{},
	}
}

func (cache directionalSchemaDiffCache) get(d direction, schema1, schema2 *openapi3.SchemaRef) (*SchemaDiff, bool) {
	if d == directionRequest {
		diff, ok := cache.requestCache[schemaPair{schema1, schema2}]
		return diff, ok
	}

	diff, ok := cache.responseCache[schemaPair{schema1, schema2}]
	return diff, ok
}

func (cache directionalSchemaDiffCache) add(d direction, schema1, schema2 *openapi3.SchemaRef, diff *SchemaDiff) {
	if d == directionRequest {
		cache.requestCache[schemaPair{schema1, schema2}] = diff
		return
	}
	cache.responseCache[schemaPair{schema1, schema2}] = diff
}
