package diff

import "github.com/getkin/kin-openapi/openapi3"

// getPrefixItemsDiff compares prefixItems by index: the schema at index i of
// prefixItems validates the array element at index i, so entry i on each side
// is a pair and is diffed as one. An entry past the end of the shorter list is
// an addition or a deletion at its index.
//
// This is why prefixItems does not go through getSubschemasDiff: that
// matching is built for allOf, anyOf and oneOf, where order carries no
// meaning, so it pairs subschemas by $ref name, content and title. Under it,
// swapping two prefixItems entries pairs each with its identical twin on the
// other side and reports no change, when every position it touches now
// validates against a different schema (#1180).
func getPrefixItemsDiff(config *Config, state *state, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) (*SubschemasDiff, error) {
	if len(schemaRefs1) == 0 && len(schemaRefs2) == 0 {
		return nil, nil
	}

	result := NewSubschemasDiff()

	for i := range min(len(schemaRefs1), len(schemaRefs2)) {
		if err := result.appendModified(config, state, schemaRefs1[i], schemaRefs2[i], i, i); err != nil {
			return nil, err
		}
	}
	for i := len(schemaRefs2); i < len(schemaRefs1); i++ {
		result.appendDeleted(i, schemaRefs1[i], schemaValue(schemaRefs1[i]).Title)
	}
	for i := len(schemaRefs1); i < len(schemaRefs2); i++ {
		result.appendAdded(i, schemaRefs2[i], schemaValue(schemaRefs2[i]).Title)
	}

	if result.Empty() {
		return nil, nil
	}

	return result, nil
}
