package diff

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/utils"
)

/*
SubschemasDiff describes the changes between a pair of subschemas under AllOf, AnyOf or OneOf
[oneOf, anyOf, allOf]: https://swagger.io/docs/specification/data-models/oneof-anyof-allof-not/
[Schema Objects]: https://swagger.io/specification/#schema-object
SubschemasDiff is a combination of two diffs:

 1. Diff of referenced schemas: subschemas under AllOf, AnyOf or OneOf defined as references to schemas under components/schemas
    - schemas with the same $ref across base and revision are compared to each other and based on the result are considered as modified or unmodified
    - other schemas are considered added/deleted

 2. Diff of inline schemas: subschemas defined directly under AllOf, AnyOf or OneOf, without a reference to components/schemas
    Unlike referenced schemas, inline schemas cannot be matched by a unique name and are therefor compared by their content.
    - syntactically identical schemas across base and revision are considered unmodified
    - schemas with the same title across base and revision are compared to each other and based on the result are considered as modified or unmodified
    - other schemas are considered added/deleted

Special case:
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

	return diffRefs.combine(diffInline)
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
		if !filter(schemaRef1) {
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
