package nullability

import "slices"

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
