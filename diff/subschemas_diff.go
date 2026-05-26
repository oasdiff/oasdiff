package diff

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

/*
SubschemasDiff describes the changes between a pair of subschemas under AllOf, AnyOf or OneOf
[oneOf, anyOf, allOf]: https://swagger.io/docs/specification/data-models/oneof-anyof-allof-not/
[Schema Objects]: https://swagger.io/specification/#schema-object
SubschemasDiff is a combination of three diffs:

 1. Diff of referenced schemas: subschemas under AllOf, AnyOf or OneOf defined as references to schemas under components/schemas
    - schemas with the same $ref across base and revision are compared to each other and based on the result are considered as modified or unmodified
    - other schemas are considered added/deleted

 2. Diff of inline schemas: subschemas defined directly under AllOf, AnyOf or OneOf, without a reference to components/schemas
    Unlike referenced schemas, inline schemas cannot be matched by a unique name and are therefor compared by their content.
    - syntactically identical schemas across base and revision are considered unmodified
    - schemas with the same title across base and revision are compared to each other and based on the result are considered as modified or unmodified
    - other schemas are considered added/deleted

 3. Reconciliation of inline/$ref refactors (AnyOf and OneOf only, when Config.MatchInlineRefs is true):
    after passes 1 and 2, any unmatched Added/Deleted pair that crosses the inline/$ref boundary and is validation-equivalent (annotation-only differences ignored) is paired and removed from Added and Deleted. Matching is pair-based: each Deleted matches at most one Added.

Special case (in pass 2):
If there remains exactly one added schema and one deleted schema without a reference and without a title, they will be be compared to eachother and considered as modified or unmodified
*/
type SubschemasDiff struct {
	Added    Subschemas         `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  Subschemas         `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedSubschemas `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// NewSubschemasDiff creates a new SubschemasDiff
func NewSubschemasDiff() *SubschemasDiff {
	return &SubschemasDiff{
		Added:    Subschemas{},
		Deleted:  Subschemas{},
		Modified: ModifiedSubschemas{},
	}
}

func (diff *SubschemasDiff) appendAdded(index int, schemaRef *openapi3.SchemaRef, title string) {
	diff.Added = append(diff.Added,
		Subschema{
			Index:     index,
			Component: getComponentName(schemaRef),
			Title:     title,
		},
	)
}

func (diff *SubschemasDiff) appendDeleted(index int, schemaRef *openapi3.SchemaRef, title string) {
	diff.Deleted = append(diff.Deleted,
		Subschema{
			Index:     index,
			Component: getComponentName(schemaRef),
			Title:     title,
		},
	)
}

func (diff *SubschemasDiff) appendModified(config *Config, state *state, schemaRef1, schemaRef2 *openapi3.SchemaRef, index1, index2 int) error {
	var err error
	diff.Modified, err = diff.Modified.addSchemaDiff(config, state, schemaRef1, schemaRef2, index1, index2)
	if err != nil {
		return err
	}
	return nil
}

func getComponentName(schemaRef *openapi3.SchemaRef) string {
	return schemaRef.Ref[strings.LastIndex(schemaRef.Ref, "/")+1:]
}

// Empty indicates whether a change was found in this element
func (diff *SubschemasDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func getSubschemasDiff(config *Config, state *state, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) (*SubschemasDiff, error) {
	diff, err := getSubschemasDiffInternal(config, state, schemaRefs1, schemaRefs2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func (diff SubschemasDiff) combine(other SubschemasDiff) (*SubschemasDiff, error) {

	return &SubschemasDiff{
		Added:    append(diff.Added, other.Added...),
		Deleted:  append(diff.Deleted, other.Deleted...),
		Modified: append(diff.Modified, other.Modified...),
	}, nil
}

func getSubschemasDiffInternal(config *Config, state *state, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) (*SubschemasDiff, error) {

	if len(schemaRefs1) == 0 && len(schemaRefs2) == 0 {
		return nil, nil
	}

	diffRefs, err := getSubschemasRefDiff(config, state, schemaRefs1, schemaRefs2)
	if err != nil {
		return nil, err
	}

	diffInline, err := getSubschemasInlineDiff(config, state, schemaRefs1, schemaRefs2)
	if err != nil {
		return nil, err
	}

	combined, err := diffRefs.combine(diffInline)
	if err != nil {
		return nil, err
	}

	return reconcileInlineRefRefactors(config, combined, schemaRefs1, schemaRefs2), nil
}

// reconcileInlineRefRefactors pairs unmatched Added and Deleted entries that
// cross the inline/$ref boundary and are validation-equivalent under the
// caller's config, then removes both members of each pair. Each Deleted
// matches at most one Added (first-fit), so a standalone addition that
// happens to duplicate an existing still-present branch is preserved.
//
// Gated on config.MatchInlineRefs; default true.
func reconcileInlineRefRefactors(config *Config, combined *SubschemasDiff, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) *SubschemasDiff {
	if !config.MatchInlineRefs {
		return combined
	}
	if combined.Empty() {
		return combined
	}

	matchedAdded := map[int]bool{}
	matchedDeleted := map[int]bool{}

	for di, deletedSubschema := range combined.Deleted {
		// Defensive: Subschema.Index is set from a walk over schemaRefs1 by the
		// earlier passes, so out-of-range is an upstream invariant violation.
		// Skip rather than panic; the entry stays in combined unchanged.
		if deletedSubschema.Index < 0 || deletedSubschema.Index >= len(schemaRefs1) {
			continue
		}
		deletedRef := schemaRefs1[deletedSubschema.Index]

		for ai, addedSubschema := range combined.Added {
			if matchedAdded[ai] {
				continue
			}
			// Same defensive guard against an upstream invariant violation,
			// this time for the revision-side index.
			if addedSubschema.Index < 0 || addedSubschema.Index >= len(schemaRefs2) {
				continue
			}
			addedRef := schemaRefs2[addedSubschema.Index]

			// Skip inline-to-inline and $ref-to-$ref pairs because the earlier
			// passes already tried byte-identity (inline) and ref-name match
			// ($ref), and failed for substantive reasons. A rename like
			// UserRoleV1 -> UserRole is still a change the user wants to see.
			if !isInlineRefactorBoundary(deletedRef, addedRef) {
				continue
			}
			if !SchemaRefsValidationEquivalent(config, deletedRef, addedRef) {
				continue
			}

			matchedAdded[ai] = true
			matchedDeleted[di] = true
			break
		}
	}

	if len(matchedAdded) == 0 {
		return combined
	}

	result := &SubschemasDiff{
		Added:    make(Subschemas, 0, len(combined.Added)-len(matchedAdded)),
		Deleted:  make(Subschemas, 0, len(combined.Deleted)-len(matchedDeleted)),
		Modified: combined.Modified,
	}
	for ai, addedSubschema := range combined.Added {
		if matchedAdded[ai] {
			continue
		}
		result.Added = append(result.Added, addedSubschema)
	}
	for di, deletedSubschema := range combined.Deleted {
		if matchedDeleted[di] {
			continue
		}
		result.Deleted = append(result.Deleted, deletedSubschema)
	}

	return result
}

// isInlineRefactorBoundary reports whether the two refs sit on opposite sides
// of the inline/$ref boundary. An inline-to-inline edit (or $ref-to-$ref
// rename) is intentionally out of scope: only the cross-boundary refactor is
// recognised as form-only.
func isInlineRefactorBoundary(schemaRef1, schemaRef2 *openapi3.SchemaRef) bool {
	if schemaRef1 == nil || schemaRef2 == nil {
		return false
	}
	return (schemaRef1.Ref == "") != (schemaRef2.Ref == "")
}

type schemaRefsFilter func(schemaRef *openapi3.SchemaRef) bool

// getSubschemasRefDiff compares subschemas by $ref name
func getSubschemasRefDiff(config *Config, state *state, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) (*SubschemasDiff, error) {

	result := NewSubschemasDiff()

	refMap2 := toRefMap(schemaRefs2, isSchemaRef)
	for index1, schemaRef1 := range schemaRefs1 {
		if !isSchemaRef(schemaRef1) {
			continue
		}
		if schemaRef2, index2, found := refMap2.pop(schemaRef1.Ref); found {
			if err := result.appendModified(config, state, schemaRef1, schemaRef2, index1, index2); err != nil {
				return result, err
			}
			continue
		}
		result.appendDeleted(index1, schemaRef1, "")
	}

	refMap1 := toRefMap(schemaRefs1, isSchemaRef)
	for index2, schemaRef2 := range schemaRefs2 {
		if !isSchemaRef(schemaRef2) {
			continue
		}
		if _, _, found := refMap1.pop(schemaRef2.Ref); !found {
			result.appendAdded(index2, schemaRef2, "")
		}
	}
	return result, nil
}

// getSubschemasInlineDiff compares inline subschemas
func getSubschemasInlineDiff(config *Config, state *state, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) (SubschemasDiff, error) {

	// find schemas in revision that have no matching schema in the base
	addedIdx, err := getNonContainedInlineSchemas(config, state, schemaRefs2, schemaRefs1)
	if err != nil {
		return SubschemasDiff{}, err
	}

	// find schemas in base that have no matching schema in the revision
	deletedIdx, err := getNonContainedInlineSchemas(config, state, schemaRefs1, schemaRefs2)
	if err != nil {
		return SubschemasDiff{}, err
	}

	// match schemas by title
	addedIdx, deletedIdx, modifiedSchemas, err := compareByTitle(config, state, addedIdx, deletedIdx, schemaRefs1, schemaRefs2)
	if err != nil {
		return SubschemasDiff{}, err
	}

	// special case: single modified schema with no title
	if isSingleModifiedCase(schemaRefs1, schemaRefs2, addedIdx, deletedIdx) {
		var err error
		modifiedSchemas, err = modifiedSchemas.addSchemaDiff(config, state, schemaRefs1[deletedIdx[0]], schemaRefs2[addedIdx[0]], deletedIdx[0], addedIdx[0])
		if err != nil {
			return SubschemasDiff{}, err
		}
		addedIdx = []int{}
		deletedIdx = []int{}
	}

	return SubschemasDiff{
		Added:    getSubschemas(addedIdx, schemaRefs2),
		Deleted:  getSubschemas(deletedIdx, schemaRefs1),
		Modified: modifiedSchemas,
	}, nil
}

func isSingleModifiedCase(schemaRefs1, schemaRefs2 openapi3.SchemaRefs, addedIdx, deletedIdx []int) bool {
	return len(addedIdx) == 1 &&
		len(deletedIdx) == 1 &&
		schemaRefs1[deletedIdx[0]].Value.Title == "" &&
		schemaRefs2[addedIdx[0]].Value.Title == ""
}

func compareByTitle(config *Config, state *state, addedIdx, deletedIdx []int, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) ([]int, []int, ModifiedSubschemas, error) {

	addedMatched, deletedMatched := matchByTitle(addedIdx, deletedIdx, schemaRefs1, schemaRefs2)

	modifiedSchemas := ModifiedSubschemas{}
	for _, addedId := range addedIdx {
		deletedId, found := addedMatched[addedId]
		if !found {
			continue
		}

		var err error
		modifiedSchemas, err = modifiedSchemas.addSchemaDiff(config, state, schemaRefs1[deletedId], schemaRefs2[addedId], deletedId, addedId)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return deleteMatched(addedIdx, addedMatched), deleteMatched(deletedIdx, deletedMatched), modifiedSchemas, nil
}

func deleteMatched(idx []int, addedMatched map[int]int) []int {
	addedIdxRemaining := []int{}
	for _, addedId := range idx {
		if _, found := addedMatched[addedId]; !found {
			addedIdxRemaining = append(addedIdxRemaining, addedId)
		}
	}
	return addedIdxRemaining
}

func matchByTitle(addedIdx, deletedIdx []int, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) (map[int]int, map[int]int) {

	addedMatched := map[int]int{}
	deletedMatched := map[int]int{}

	matchedTitles := utils.StringSet{}
	matchedTitles.Add("") // empty title is not allowed

	for _, addedId := range addedIdx {
		title := schemaRefs2[addedId].Value.Title
		if matchedTitles.Contains(title) {
			// title already matched, skip
			continue
		}
		for _, deletedId := range deletedIdx {
			if title == schemaRefs1[deletedId].Value.Title {
				addedMatched[addedId] = deletedId
				deletedMatched[deletedId] = addedId
				matchedTitles.Add(title)
				break
			}
		}
	}

	return addedMatched, deletedMatched
}

func getNonContainedInlineSchemas(config *Config, state *state, schemaRefs1, schemaRefs2 openapi3.SchemaRefs) ([]int, error) {

	notContainedIdx := []int{}
	matched := map[int]struct{}{}

	for index1, schemaRef1 := range schemaRefs1 {
		if !isSchemaInline(schemaRef1) {
			continue
		}

		if found, index2, err := findIndenticalSchema(config, state, schemaRef1, schemaRefs2, matched, isSchemaInline); err != nil {
			return nil, err
		} else if !found {
			notContainedIdx = append(notContainedIdx, index1)
		} else {
			matched[index2] = struct{}{}
		}
	}
	return notContainedIdx, nil
}

func findIndenticalSchema(config *Config, state *state, schemaRef1 *openapi3.SchemaRef, schemasRefs2 openapi3.SchemaRefs, matched map[int]struct{}, filter schemaRefsFilter) (bool, int, error) {
	for index2, schemaRef2 := range schemasRefs2 {
		// Restrict candidates to those matching the filter. schemaRef1 is
		// the caller's already-filtered source; only the candidate side
		// needs filtering here.
		if !filter(schemaRef2) {
			continue
		}
		if alreadyMatched(index2, matched) {
			continue
		}

		if schemaDiff, err := getSchemaDiff(config, state, schemaRef1, schemaRef2); err != nil {
			return false, 0, err
		} else if schemaDiff.Empty() {
			return true, index2, nil
		}
	}

	return false, 0, nil
}

func alreadyMatched(index int, matched map[int]struct{}) bool {
	_, found := matched[index]
	return found
}

func isSchemaInline(schemaRef *openapi3.SchemaRef) bool {
	if schemaRef == nil {
		return false
	}
	return schemaRef.Ref == ""
}

func isSchemaRef(schemaRef *openapi3.SchemaRef) bool {
	if schemaRef == nil {
		return false
	}
	return schemaRef.Ref != ""
}
