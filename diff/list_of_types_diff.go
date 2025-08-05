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
	Added   []string `json:"added,omitempty" yaml:"added,omitempty"`   // Types added to the list-of-types pattern
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
	
	return compareTypePatterns(basePattern, revisionPattern)
}

// detectTypePattern identifies types from single type schemas OR list-of-types patterns.
// It detects the "list-of-types" design pattern where oneOf/anyOf contain only simple type schemas.
// Returns nil if the schema doesn't match this pattern (e.g., complex objects, nested schemas).
func detectTypePattern(schema *openapi3.Schema) *typePattern {
	if schema == nil {
		return nil
	}
	
	// Check for list-of-types patterns first (oneOf/anyOf)
	if pattern := checkSchemaList(schema.OneOf, "oneOf"); pattern != nil {
		return &typePattern{Types: pattern.Types, Source: "list"}
	}
	
	if pattern := checkSchemaList(schema.AnyOf, "anyOf"); pattern != nil {
		return &typePattern{Types: pattern.Types, Source: "list"}
	}
	
	// Check for single type
	if schema.Type != nil && len(*schema.Type) == 1 {
		return &typePattern{
			Types:  []string{(*schema.Type)[0]},
			Source: "single",
		}
	}
	
	return nil
}

// listOfTypesPattern represents a detected list-of-types pattern
type listOfTypesPattern struct {
	Types  []string // e.g., ["string", "integer"]
	Source string   // "oneOf" or "anyOf"
}

// checkSchemaList determines if a oneOf/anyOf schema list represents the list-of-types pattern.
// Returns non-nil only if ALL schemas in the list are simple type schemas (no complex objects).
func checkSchemaList(schemas []*openapi3.SchemaRef, source string) *listOfTypesPattern {
	if len(schemas) == 0 {
		return nil
	}
	
	var types []string
	for _, schemaRef := range schemas {
		if schemaRef.Value == nil {
			return nil // Can't analyze refs or nil schemas
		}
		
		schema := schemaRef.Value
		
		// Must be simple type (single type, no complex properties)
		if !isSimpleTypeSchema(schema) {
			return nil
		}
		
		types = append(types, getSchemaType(schema))
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
// Simple types have exactly one type and no complex properties, oneOf, anyOf, or allOf.
func isSimpleTypeSchema(schema *openapi3.Schema) bool {
	// Must have exactly one type
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