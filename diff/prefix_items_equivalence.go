package diff

import "github.com/getkin/kin-openapi/openapi3"

// PrefixItemsValidationEquivalent reports whether two schemas validate every
// array position covered by prefixItems against the same contract. A position
// past the end of prefixItems is governed by items, so adding or removing an
// entry leaves the accepted arrays unchanged when the entry and the items
// schema it stands in for have the same contract.
func PrefixItemsValidationEquivalent(config *Config, base, revision *openapi3.Schema) bool {
	if base == nil || revision == nil {
		return false
	}

	for i := range max(len(base.PrefixItems), len(revision.PrefixItems)) {
		if !SchemaRefsValidationEquivalent(config, itemSchemaAt(base, i), itemSchemaAt(revision, i)) {
			return false
		}
	}
	return true
}

// itemSchemaAt returns the schema an array element at index i is validated
// against: its prefixItems entry, or items past the end of the list, or the
// empty schema when items is absent and the position is unconstrained.
func itemSchemaAt(schema *openapi3.Schema, i int) *openapi3.SchemaRef {
	if i < len(schema.PrefixItems) {
		return schema.PrefixItems[i]
	}
	if schema.Items == nil {
		return emptySchemaRef()
	}
	return schema.Items
}
