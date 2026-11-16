package nullability

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

const nullType = "null"

// Transition is how a schema's acceptance of null changed.
type Transition int

const (
	Unchanged Transition = iota
	BecameNullable
	BecameNotNullable
)

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
// constrain the instance alongside the composition.
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

// ComposedChange compares what ComposedAccepts says of the two versions of a
// schema. It is Unchanged when either cannot be decided.
func ComposedChange(base, revision *openapi3.Schema) Transition {
	before, beforeDecided := ComposedAccepts(base)
	after, afterDecided := ComposedAccepts(revision)
	switch {
	case !beforeDecided || !afterDecided || before == after:
		return Unchanged
	case after:
		return BecameNullable
	default:
		return BecameNotNullable
	}
}

// AddedToTypes reports whether "null" was added to a type array that already
// constrained the type. A schema with no type keyword accepted every value,
// null included, so giving it a type array that contains null does not make
// it newly nullable.
func AddedToTypes(added []string, baseTypes []string) bool {
	return len(baseTypes) > 0 && slices.Contains(added, nullType)
}

// RemovedFromTypes reports whether "null" was removed from a type array that
// still constrains the type. A schema whose type keyword was dropped accepts
// every value, null included, so it did not become non-nullable.
func RemovedFromTypes(deleted []string, revisionTypes []string) bool {
	return len(revisionTypes) > 0 && slices.Contains(deleted, nullType)
}

// OnlyNullChanged reports whether adding or removing "null" is the only
// change to a type array, given the types added and deleted.
func OnlyNullChanged(added, deleted []string) bool {
	return onlyNull(added) && onlyNull(deleted)
}

func onlyNull(types []string) bool {
	return !slices.ContainsFunc(types, func(t string) bool { return t != nullType })
}

// WithoutNull returns the types with "null" removed, so that a nullable type
// such as ["string", "null"] compares by its non-null types.
func WithoutNull(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		if t != nullType {
			out = append(out, t)
		}
	}
	return out
}
