package allof

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// MergeSpec merges all instances of allOf in place, across every schema in
// the document. Merge handles each schema's whole subtree (including its own
// cycle tracking), so the walk hands it each attachment point once and skips
// descent.
func MergeSpec(spec *openapi3.T) (*openapi3.T, error) {
	err := spec.WalkSchemas(func(_ string, s *openapi3.SchemaRef) error {
		m, err := Merge(*s)
		if err != nil {
			return err
		}
		// Every $ref to this schema shares one Value, so writing the merge
		// into it updates every use. Assigning s.Value would update only
		// this reference and leave the rest unmerged.
		*s.Value = *m
		return openapi3.SkipSubtree
	})
	return spec, err
}
