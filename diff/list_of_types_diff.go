package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ListOfTypesDiff represents changes in the "list-of-types" design pattern.
// This is not an OpenAPI object, but rather a common pattern where oneOf/anyOf
// schemas are used with simple type schemas to allow multiple types for a field.
// For example: oneOf: [{type: string}, {type: integer}] allows string OR integer values.
// This diff tracks when types are added to or removed from such patterns.
type ListOfTypesDiff struct {
	Added   []string `json:"added,omitempty" yaml:"added,omitempty"`     // Types added to the list-of-types pattern
	Deleted []string `json:"deleted,omitempty" yaml:"deleted,omitempty"` // Types removed from the list-of-types pattern
}

// Empty indicates whether a change was found in this element
func (diff *ListOfTypesDiff) Empty() bool {
	return diff == nil || (len(diff.Added) == 0 && len(diff.Deleted) == 0)
}

// typePattern represents a detected type pattern from a schema
type typePattern struct {
	Types  []string // e.g., ["string"] or ["string", "integer"]
	Source string   // "single" or "list"
}

// getListOfTypesDiff detects and compares the "list-of-types" design pattern.
// It handles transitions between single types and oneOf/anyOf patterns that represent
// multiple allowed types. For example:
//   - type: string -> oneOf: [{type: string}, {type: integer}] (integer added)
//   - anyOf: [{type: string}, {type: number}] -> type: string (number deleted)
//   - oneOf: [{type: string}] -> anyOf: [{type: string}, {type: boolean}] (boolean added)
func getListOfTypesDiff(base, revision *openapi3.Schema) *ListOfTypesDiff {
	basePattern := detectTypePattern(base)
	revisionPattern := detectTypePattern(revision)

	// Only create diff if patterns are different
	if areTypePatternsEqual(basePattern, revisionPattern) {
		return nil
	}

	// Only create ListOfTypesDiff when:
	// 1. Both patterns are valid (non-nil) - all schemas must have exactly one type
	// 2. At least one side involves a list pattern (oneOf/anyOf)
	// This ensures we only handle transitions between valid type patterns
	if basePattern == nil || revisionPattern == nil {
		return nil // Invalid pattern - contains schema without type
	}

	hasListPattern := (basePattern.Source == "list") || (revisionPattern.Source == "list")
	if !hasListPattern {
		return nil
	}

	return compareTypePatterns(basePattern, revisionPattern)
}

// detectTypePattern identifies types from single type schemas OR list-of-types patterns.
// It detects the "list-of-types" design pattern where oneOf/anyOf contain only simple type schemas.
// Returns nil if the schema doesn't match this pattern (e.g., complex objects, nested schemas).
func detectTypePattern(schema *openapi3.Schema) *typePattern {
	if schema == nil {
		return nil
	}

	// Exclude schemas with complex constructs from list-of-types detection
	// These should be handled by other diff mechanisms
	if schema.Not != nil || len(schema.AllOf) > 0 || len(schema.Properties) > 0 {
		return nil
	}

	// Check for single type first - if a schema has both type and empty oneOf/anyOf,
	// the type should take precedence (common when schemas are being transformed)
	if schema.Type != nil && len(*schema.Type) >= 1 {
		return &typePattern{
			// Take the first type value - while the kin-openapi library models Type as []string,
			// the OpenAPI 3.0 specification defines type as a single string value.
			// For multiple types, OpenAPI 3.0 uses oneOf/anyOf instead of type arrays.
			Types:  []string{(*schema.Type)[0]},
			Source: "single",
		}
	}

	// Check for list-of-types patterns (oneOf/anyOf)
	// Only analyze as list-of-types if ALL schemas have exactly one type
	oneOfPattern := checkSchemaList(schema.OneOf, "oneOf")
	anyOfPattern := checkSchemaList(schema.AnyOf, "anyOf")

	// Use oneOf if it has types, otherwise use anyOf if it has types
	if oneOfPattern != nil && len(oneOfPattern.Types) > 0 {
		return &typePattern{Types: oneOfPattern.Types, Source: "list"}
	}
	if anyOfPattern != nil && len(anyOfPattern.Types) > 0 {
		return &typePattern{Types: anyOfPattern.Types, Source: "list"}
	}

	// No list-of-types pattern found
	return nil
}

// listOfTypesPattern represents a detected list-of-types pattern
type listOfTypesPattern struct {
	Types  []string // e.g., ["string", "integer"]
	Source string   // "oneOf" or "anyOf"
}

// checkSchemaList determines if a oneOf/anyOf schema list represents the list-of-types pattern.
// Returns non-nil only if ALL schemas in the list are simple type schemas with exactly one type.
func checkSchemaList(schemas []*openapi3.SchemaRef, source string) *listOfTypesPattern {
	if len(schemas) == 0 {
		// Empty oneOf/anyOf is not a valid list-of-types pattern
		return nil
	}

	var types []string
	for _, schemaRef := range schemas {
		if schemaRef.Value == nil {
			return nil // Can't analyze refs or nil schemas
		}

		schema := schemaRef.Value

		// Must be simple type schema with exactly one type
		if !isSimpleTypeSchema(schema) {
			return nil
		}

		schemaType := getSchemaType(schema)
		if schemaType == "" {
			return nil // Schema has no type - not supported
		}

		types = append(types, schemaType)
	}

	return &listOfTypesPattern{
		Types:  types,
		Source: source,
	}
}

// compareTypePatterns generates the diff between two type patterns
func compareTypePatterns(base, revision *typePattern) *ListOfTypesDiff {
	baseTypes := make(map[string]bool)
	revisionTypes := make(map[string]bool)

	if base != nil {
		for _, t := range base.Types {
			baseTypes[t] = true
		}
	}

	if revision != nil {
		for _, t := range revision.Types {
			revisionTypes[t] = true
		}
	}

	diff := &ListOfTypesDiff{}

	// Find added types
	for t := range revisionTypes {
		if !baseTypes[t] {
			diff.Added = append(diff.Added, t)
		}
	}

	// Find deleted types
	for t := range baseTypes {
		if !revisionTypes[t] {
			diff.Deleted = append(diff.Deleted, t)
		}
	}

	if diff.Empty() {
		return nil
	}

	return diff
}

// areTypePatternsEqual checks if two type patterns are equivalent
func areTypePatternsEqual(base, revision *typePattern) bool {
	if base == nil && revision == nil {
		return true
	}
	if base == nil || revision == nil {
		return false
	}

	if len(base.Types) != len(revision.Types) {
		return false
	}

	baseSet := make(map[string]bool)
	for _, t := range base.Types {
		baseSet[t] = true
	}

	for _, t := range revision.Types {
		if !baseSet[t] {
			return false
		}
	}

	return true
}

// isSimpleTypeSchema checks if schema represents a simple type suitable for list-of-types pattern.
// Simple types must have exactly one type and no complex properties, oneOf, anyOf, or allOf.
func isSimpleTypeSchema(schema *openapi3.Schema) bool {
	// Must have exactly one type (no Any type support)
	if schema.Type == nil || len(*schema.Type) != 1 {
		return false
	}

	// Must not have complex properties
	if len(schema.Properties) > 0 ||
		len(schema.OneOf) > 0 ||
		len(schema.AnyOf) > 0 ||
		len(schema.AllOf) > 0 {
		return false
	}

	return true
}

// getSchemaType extracts the type string from a simple schema
func getSchemaType(schema *openapi3.Schema) string {
	if schema.Type != nil && len(*schema.Type) > 0 {
		return (*schema.Type)[0]
	}
	return ""
}
