package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type ref struct {
	schemaRef *openapi3.SchemaRef
	indices   []int
}

// refMap is used to match referenced subschemas across base and revision
type refMap map[string]*ref

// push adds a schema reference to the map or, if exists, adds an index to the existing reference
func (m refMap) push(schemaRef *openapi3.SchemaRef, index int) {
	if val, found := m[schemaRef.Ref]; found {
		val.indices = append(val.indices, index)
	} else {
		m[schemaRef.Ref] = &ref{
			schemaRef: schemaRef,
			indices:   []int{index},
		}
	}
}

// pop checks if the reference exists in the map, returns the schema reference and an index if exists, and removes the index from the map
func (m refMap) pop(ref string) (*openapi3.SchemaRef, int, bool) {
	if val, found := m[ref]; found {
		if len(val.indices) > 0 {
			index := val.indices[0]
			val.indices = val.indices[1:]
			return val.schemaRef, index, true
		}
		delete(m, ref)
	}
	return nil, 0, false
}

func toRefMap(schemaRefs openapi3.SchemaRefs, filter schemaRefsFilter) refMap {
	result := refMap{}
	for index, schemaRef := range schemaRefs {
		if filter(schemaRef) {
			result.push(schemaRef, index)
		}
	}
	return result
}
