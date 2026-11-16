package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/internal/nullability"
)

// ComposedNullabilityOneOfCommentId explains a verdict the branch alone
// contradicts: the branch gained null and the property stopped accepting it.
const ComposedNullabilityOneOfCommentId = "composed-nullability-one-of-comment"

// A oneOf or anyOf decides whether the schema holding it accepts null (see
// nullability.ComposedAccepts), so that schema is where the transition is
// reported (composedNullability), and a branch's own transition is not
// (branchOfComposition).

// composedNullability reports the nullability transition of a schema whose
// oneOf or anyOf decides whether it accepts null, and the comment to report it
// with.
func composedNullability(d *diff.SchemaDiff) (nullability.Transition, string) {
	transition := nullability.ComposedChange(d.Base, d.Revision)
	if transition == nullability.Unchanged {
		return transition, ""
	}
	return transition, composedNullabilityComment(d.Base, d.Revision)
}

// composedNullabilityComment explains a verdict that the changed branch alone
// contradicts: a branch gained null and the oneOf stopped accepting it, or a
// branch lost null and the oneOf started accepting it. That happens only when
// the oneOf has another nullable branch on both sides; with none, the oneOf
// follows the branch and there is nothing to explain.
func composedNullabilityComment(base, revision *openapi3.Schema) string {
	baseNullable, baseKnown := nullability.NullableBranches(base.OneOf)
	revisionNullable, revisionKnown := nullability.NullableBranches(revision.OneOf)
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
	if _, ok := nullability.ComposedAccepts(parent.Base); !ok {
		return false
	}
	if _, ok := nullability.ComposedAccepts(parent.Revision); !ok {
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
