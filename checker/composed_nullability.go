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

// ComposedNullabilityUntypedBranchCommentId explains a oneOf that rejects
// null because one of its branches has no type.
const ComposedNullabilityUntypedBranchCommentId = "composed-nullability-untyped-branch-comment"

// A oneOf or anyOf decides whether the schema holding it accepts null (see
// nullability.ComposedAccepts), so that schema is where the transition is
// reported (composedNullability), and a branch's own transition is not
// (branchOfComposition).

// composedNullability reports the nullability transition of a schema that has
// a oneOf or anyOf on either side, by comparing whether it accepted null
// before and after, and the comment to report it with. It reports whether
// both sides could be decided.
//
// With a composition on one side only, this covers a schema wrapped in
// oneOf: [{type: "null"}, X] or unwrapped from it, and decides it by what the
// two sides accept: wrapping an X that already accepts null takes null out,
// because null then matches both branches.
func composedNullability(dialects nullability.Dialects, d *diff.SchemaDiff) (nullability.Transition, string, bool) {
	if !nullability.Composed(d.Base) && !nullability.Composed(d.Revision) {
		return nullability.Unchanged, "", false
	}
	transition, decided := nullability.Change(d.Base, d.Revision, dialects)
	if transition == nullability.Unchanged {
		return transition, "", decided
	}
	return transition, composedNullabilityComment(dialects, d.Base, d.Revision), true
}

// composedNullabilityComment explains a verdict that the edit alone
// contradicts: null was added to a branch, or a null branch was added, and
// the oneOf stopped accepting null, or the reverse. That happens only when a
// oneOf has more than one branch accepting null on one of the sides; with at
// most one on both, the oneOf follows the edit and there is nothing to
// explain.
func composedNullabilityComment(dialects nullability.Dialects, base, revision *openapi3.Schema) string {
	sides := []struct {
		schema  *openapi3.Schema
		dialect nullability.Dialect
	}{{revision, dialects.Revision}, {base, dialects.Base}}
	for _, side := range sides {
		schema := side.schema
		nullable, decided := nullability.NullableBranches(schema.OneOf, side.dialect)
		if !decided || nullable < 2 {
			continue
		}
		if hasUntypedBranch(schema.OneOf, side.dialect) {
			return ComposedNullabilityUntypedBranchCommentId
		}
		return ComposedNullabilityOneOfCommentId
	}
	return ""
}

// hasUntypedBranch reports whether a branch accepts null by having no type
// keyword, which is easy to miss when reading the oneOf.
func hasUntypedBranch(branches openapi3.SchemaRefs, dialect nullability.Dialect) bool {
	return slices.ContainsFunc(branches, func(branch *openapi3.SchemaRef) bool {
		return branch != nil && branch.Value != nil &&
			len(branch.Value.Type.Slice()) == 0 && !branch.Value.Nullable &&
			nullability.OwnKeywordsAccept(branch.Value, dialect)
	})
}

// branchOfComposition reports whether d is a branch of parent's oneOf or
// anyOf, which parent reports the nullability of.
func branchOfComposition(dialects nullability.Dialects, parent *diff.SchemaDiff, d *diff.SchemaDiff) bool {
	if parent == nil || d == nil {
		return false
	}
	if _, ok := nullability.ComposedAccepts(parent.Base, dialects.Base); !ok {
		return false
	}
	if _, ok := nullability.ComposedAccepts(parent.Revision, dialects.Revision); !ok {
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
