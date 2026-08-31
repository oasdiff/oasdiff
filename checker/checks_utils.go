package checker

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

func commentId(id string) string {
	return id + "-comment"
}

func descriptionId(id string) string {
	return id + "-description"
}

func propertyFullName(propertyPath string, propertyNames ...string) string {
	propertyFullName := strings.Join(propertyNames, "/")
	if propertyPath != "" {
		propertyFullName = propertyPath + "/" + propertyFullName
	}
	return propertyFullName
}

func interfaceToString(arg any) string {
	if arg == nil {
		return "undefined"
	}

	if argString, ok := arg.(string); ok {
		return argString
	}

	if argStringList, ok := arg.([]string); ok {
		return strings.Join(argStringList, ", ")
	}

	if argInt, ok := arg.(int); ok {
		return fmt.Sprintf("%d", argInt)
	}

	if argUint64, ok := arg.(uint64); ok {
		return fmt.Sprintf("%d", argUint64)
	}

	if argFloat64, ok := arg.(float64); ok {
		return fmt.Sprintf("%.2f", argFloat64)
	}

	if argBool, ok := arg.(bool); ok {
		return fmt.Sprintf("%t", argBool)
	}

	return fmt.Sprintf("%v", arg)
}

func IsIncreased(from any, to any) bool {
	fromUint64, ok := from.(uint64)
	toUint64, okTo := to.(uint64)
	if ok && okTo {
		return fromUint64 < toUint64
	}
	fromFloat64, ok := from.(float64)
	toFloat64, okTo := to.(float64)
	if ok && okTo {
		return fromFloat64 < toFloat64
	}
	return false
}

// minItems is a value keyword whose zero is the absent constraint: minItems: 0
// requires nothing, so the diff reports absence as 0 rather than nil, and
// setting or unsetting the bound is a transition to or from zero.
func uintBoundSet(d *diff.ValueDiff) bool {
	return d != nil && d.From == uint64(0) && d.To != uint64(0)
}

func uintBoundUnset(d *diff.ValueDiff) bool {
	return d != nil && d.From != uint64(0) && d.To == uint64(0)
}

func isIncreasedValue(diff *diff.ValueDiff) bool {
	return IsIncreased(diff.From, diff.To)
}

func isDecreasedValue(diff *diff.ValueDiff) bool {
	return IsDecreased(diff.From, diff.To)
}

func IsDecreased(from any, to any) bool {
	fromUint64, ok := from.(uint64)
	toUint64, okTo := to.(uint64)
	if ok && okTo {
		return fromUint64 > toUint64
	}
	fromFloat64, ok := from.(float64)
	toFloat64, okTo := to.(float64)
	if ok && okTo {
		return fromFloat64 > toFloat64
	}
	return false
}

// splitSubschemasByAnnotationOnly partitions subschemas into two disjoint
// sets according to whether the body at Subschema.Index in schemaRefs
// carries a validation-significant keyword.
//
//   - kept: subschemas with at least one constraint, type, property,
//     enum, etc. These drive the original allOf-changed emission at its
//     conventional severity (ERR / WARN for request, INFO for response).
//   - annotationOnly: subschemas that hold only annotation keywords
//     (title, description, examples, default, externalDocs, $comment).
//     Validation-equivalent to {} — they don't reject any previously-
//     valid instance. Callers emit them at INFO so the document-level
//     change stays auditable in the changelog without contaminating
//     breaking.
//
// Motivating case: handrews on OAS discussion #3793 — adding an
// `allOf: [{title: "..."}]` is not a breaking change the way adding a
// real constraint is, but the original allOf-added check fired ERR on it.
func splitSubschemasByAnnotationOnly(subschemas diff.Subschemas, schemaRefs openapi3.SchemaRefs) (kept diff.Subschemas, annotationOnly diff.Subschemas) {
	if len(subschemas) == 0 {
		return subschemas, nil
	}
	kept = make(diff.Subschemas, 0, len(subschemas))
	annotationOnly = make(diff.Subschemas, 0, len(subschemas))
	emptyRef := &openapi3.SchemaRef{Value: &openapi3.Schema{}}
	for _, s := range subschemas {
		isAnnotationOnly := false
		if s.Index >= 0 && s.Index < len(schemaRefs) {
			ref := schemaRefs[s.Index]
			if ref != nil && ref.Value != nil &&
				diff.SchemaRefsValidationEquivalent(diff.NewConfig(), emptyRef, ref) {
				isAnnotationOnly = true
			}
		}
		if isAnnotationOnly {
			annotationOnly = append(annotationOnly, s)
		} else {
			kept = append(kept, s)
		}
	}
	return kept, annotationOnly
}

// prefixItemsChangedContract reports whether a prefixItems diff changed which
// arrays the schema accepts. An entry added or removed at a position the items
// schema governs identically leaves the accepted arrays untouched, so the
// syntactic diff alone is not a change to report.
func prefixItemsChangedContract(schemaDiff *diff.SchemaDiff) bool {
	return schemaDiff.PrefixItemsDiff != nil &&
		!diff.PrefixItemsValidationEquivalent(diff.NewConfig(), schemaDiff.Base, schemaDiff.Revision)
}
