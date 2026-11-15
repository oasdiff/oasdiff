package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// ComposedNullabilityOneOfCommentId explains a verdict the branch alone
// contradicts: the branch gained null and the property stopped accepting it.
const ComposedNullabilityOneOfCommentId = "composed-nullability-one-of-comment"

// A oneOf matches an instance only when exactly one of its branches does, so a
// oneOf accepts null only while exactly one of its branches accepts null:
// making a second branch nullable takes null out of what the schema accepts,
// and making one of two nullable branches non-nullable puts it back. An anyOf
// accepts null when any branch does, so a branch's nullability matters there
// only while it is the only nullable branch.
//
// Either way the branch does not decide on its own, so the schema holding the
// composition is where the transition is reported (composedNullability), and
// the branch's own transition is not (branchOfComposition).

// composedNullability reports the nullability transition of a schema whose
// oneOf or anyOf decides whether it accepts null, and the comment to report it
// with. It returns nullabilityUnchanged when either side cannot be decided.
func composedNullability(d *diff.SchemaDiff) (nullability, string) {
	base, baseKnown := composedAcceptsNull(d.Base)
	revision, revisionKnown := composedAcceptsNull(d.Revision)
	if !baseKnown || !revisionKnown || base == revision {
		return nullabilityUnchanged, ""
	}
	comment := composedNullabilityComment(d.Base, d.Revision)
	if revision {
		return becameNullable, comment
	}
	return becameNotNullable, comment
}

// composedNullabilityComment explains a verdict that the changed branch alone
// contradicts: a branch gained null and the oneOf stopped accepting it, or a
// branch lost null and the oneOf started accepting it. That happens only when
// the oneOf has another nullable branch on both sides; with none, the oneOf
// follows the branch and there is nothing to explain.
func composedNullabilityComment(base, revision *openapi3.Schema) string {
	baseNullable, baseKnown := nullableBranches(base.OneOf)
	revisionNullable, revisionKnown := nullableBranches(revision.OneOf)
	if !baseKnown || !revisionKnown {
		return ""
	}
	if baseNullable > 0 && revisionNullable > 0 {
		return ComposedNullabilityOneOfCommentId
	}
	return ""
}

// branchOfComposition reports whether d is a branch of parent's oneOf or
// anyOf, which parent reports the nullability of.
func branchOfComposition(parent *diff.SchemaDiff, d *diff.SchemaDiff) bool {
	if parent == nil || d == nil {
		return false
	}
	if _, ok := composedAcceptsNull(parent.Base); !ok {
		return false
	}
	if _, ok := composedAcceptsNull(parent.Revision); !ok {
		return false
	}
	return isBranch(parent.OneOfDiff, d) || isBranch(parent.AnyOfDiff, d)
}

func isBranch(subschemas *diff.SubschemasDiff, d *diff.SchemaDiff) bool {
	if subschemas == nil {
		return false
	}
	return slices.ContainsFunc(subschemas.Modified, func(modified *diff.ModifiedSubschema) bool {
		return modified != nil && modified.Diff == d
	})
}

// composedAcceptsNull reports whether a schema accepts null, for a schema
// whose oneOf or anyOf decides that, and whether it could be decided.
func composedAcceptsNull(schema *openapi3.Schema) (bool, bool) {
	if schema == nil {
		return false, false
	}
	if len(schema.OneOf) == 0 && len(schema.AnyOf) == 0 {
		return false, false
	}
	// allOf and not constrain the instance alongside the composition, and
	// what they accept is not decided here.
	if len(schema.AllOf) > 0 || schema.Not != nil {
		return false, false
	}

	accepts := typePermitsNull(schema)
	if len(schema.OneOf) > 0 {
		nullable, ok := nullableBranches(schema.OneOf)
		if !ok {
			return false, false
		}
		accepts = accepts && nullable == 1
	}
	if len(schema.AnyOf) > 0 {
		nullable, ok := nullableBranches(schema.AnyOf)
		if !ok {
			return false, false
		}
		accepts = accepts && nullable > 0
	}
	return accepts, true
}

// nullableBranches counts the branches that accept null, and reports whether
// every branch could be decided: a branch that composes further is not,
// because diff.SchemaAcceptsNull reads a schema's own type and nullable.
func nullableBranches(branches openapi3.SchemaRefs) (int, bool) {
	count := 0
	for _, branch := range branches {
		if branch == nil || branch.Value == nil {
			return 0, false
		}
		schema := branch.Value
		if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 || len(schema.AllOf) > 0 || schema.Not != nil {
			return 0, false
		}
		if diff.SchemaAcceptsNull(schema) {
			count++
		}
	}
	return count, true
}

// typePermitsNull reports whether a schema's own keywords let null through. A
// schema with no type keyword constrains nothing, so it does.
func typePermitsNull(schema *openapi3.Schema) bool {
	if schema.Nullable {
		return true
	}
	types := schema.Type.Slice()
	if len(types) == 0 {
		return true
	}
	return slices.Contains(types, "null")
}
