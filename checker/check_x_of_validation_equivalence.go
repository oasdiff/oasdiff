package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

func filterValidationEquivalentDeletedSubschemas(deleted diff.Subschemas, baseRefs, revisionRefs openapi3.SchemaRefs) diff.Subschemas {
	result := diff.Subschemas{}
	for _, deletedSubschema := range deleted {
		if deletedSubschema.Index < 0 || deletedSubschema.Index >= len(baseRefs) {
			result = append(result, deletedSubschema)
			continue
		}

		if hasValidationEquivalentSubschema(baseRefs[deletedSubschema.Index], revisionRefs) {
			continue
		}

		result = append(result, deletedSubschema)
	}

	return result
}

func hasValidationEquivalentSubschema(schemaRef *openapi3.SchemaRef, schemaRefs openapi3.SchemaRefs) bool {
	// This is O(deleted x revision); x-of lists are typically short, and the
	// conservative inline/$ref boundary keeps the extra diffing scoped.
	for _, candidateRef := range schemaRefs {
		if !isInlineRefactor(schemaRef, candidateRef) {
			continue
		}
		if diff.SchemaRefsValidationEquivalent(schemaRef, candidateRef) {
			return true
		}
	}

	return false
}

func isInlineRefactor(schemaRef1, schemaRef2 *openapi3.SchemaRef) bool {
	if schemaRef1 == nil || schemaRef2 == nil {
		return false
	}

	return (schemaRef1.Ref == "") != (schemaRef2.Ref == "")
}
