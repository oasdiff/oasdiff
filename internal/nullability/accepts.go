package nullability

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

const nullType = "null"

// TypeAccepts reports whether the schema names null as an accepted type: the
// nullable keyword (OpenAPI 3.0) or a "null" entry in its type array
// (OpenAPI 3.1).
func TypeAccepts(schema *openapi3.Schema) bool {
	return schema.Nullable || slices.Contains(schema.Type.Slice(), nullType)
}

// OwnKeywordsAccept reports whether the type, nullable, enum and const of a
// schema let null through. A schema with no type keyword accepts every type,
// null included. An enum lets null through only when null is one of its
// values, and a const only when it is null.
func OwnKeywordsAccept(schema *openapi3.Schema) bool {
	if len(schema.Type.Slice()) > 0 && !TypeAccepts(schema) {
		return false
	}
	if len(schema.Enum) > 0 && !slices.ContainsFunc(schema.Enum, func(value any) bool { return value == nil }) {
		return false
	}
	return schema.Const == nil
}

// A oneOf matches an instance only when exactly one of its branches does, so a
// oneOf accepts null only while exactly one of its branches accepts null:
// making a second branch nullable takes null out of what the schema accepts,
// and making one of two nullable branches non-nullable puts it back. An anyOf
// accepts null when any branch does.

// ComposedAccepts reports whether a schema accepts null, for a schema whose
// oneOf or anyOf decides that, and whether it could be decided. It is not
// decided for a schema with neither, or with an allOf, a not or an if, which
// constrain the instance alongside the composition. A schema with
// nullable: true accepts null whatever its branches do.
func ComposedAccepts(schema *openapi3.Schema) (accepts bool, decided bool) {
	if schema == nil {
		return false, false
	}
	if len(schema.OneOf) == 0 && len(schema.AnyOf) == 0 {
		return false, false
	}
	if len(schema.AllOf) > 0 || schema.Not != nil || schema.If != nil {
		return false, false
	}
	// OpenAPI 3.0 ties the nullable keyword to a type in the same schema, so
	// beside a oneOf or anyOf it strictly does nothing. It is still how a 3.0
	// document says the composition is nullable, and it is read that way.
	if schema.Nullable {
		return true, true
	}

	accepts = OwnKeywordsAccept(schema)
	if len(schema.OneOf) > 0 {
		nullable, ok := NullableBranches(schema.OneOf)
		if !ok {
			return false, false
		}
		accepts = accepts && nullable == 1
	}
	if len(schema.AnyOf) > 0 {
		nullable, ok := NullableBranches(schema.AnyOf)
		if !ok {
			return false, false
		}
		accepts = accepts && nullable > 0
	}
	return accepts, true
}

// NullableBranches counts the branches that accept null, and reports whether
// every branch could be decided: a branch that composes further, or has a not
// or an if, is not, because OwnKeywordsAccept reads only the keywords that
// decide on their own.
func NullableBranches(branches openapi3.SchemaRefs) (count int, decided bool) {
	for _, branch := range branches {
		if branch == nil || branch.Value == nil {
			return 0, false
		}
		schema := branch.Value
		if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 || len(schema.AllOf) > 0 || schema.Not != nil || schema.If != nil {
			return 0, false
		}
		if OwnKeywordsAccept(schema) {
			count++
		}
	}
	return count, true
}

// Composed reports whether the schema has a oneOf or an anyOf.
func Composed(schema *openapi3.Schema) bool {
	return schema != nil && (len(schema.OneOf) > 0 || len(schema.AnyOf) > 0)
}

// Accepts reports whether a schema accepts null, and whether it could be
// decided: by its oneOf or anyOf when it has one (see ComposedAccepts), and
// by its own keywords otherwise. It is not decided for a schema with an
// allOf, a not or an if.
func Accepts(schema *openapi3.Schema) (accepts bool, decided bool) {
	if schema == nil {
		return false, false
	}
	if Composed(schema) {
		return ComposedAccepts(schema)
	}
	if len(schema.AllOf) > 0 || schema.Not != nil || schema.If != nil {
		return false, false
	}
	return OwnKeywordsAccept(schema), true
}
