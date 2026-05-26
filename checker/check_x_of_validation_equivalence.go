package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

func filterValidationEquivalentDeletedSubschemas(deleted diff.Subschemas, baseRefs, revisionRefs openapi3.SchemaRefs) diff.Subschemas {
	return filterValidationEquivalentSubschemas(deleted, baseRefs, revisionRefs)
}

func filterValidationEquivalentAddedSubschemas(added diff.Subschemas, baseRefs, revisionRefs openapi3.SchemaRefs) diff.Subschemas {
	return filterValidationEquivalentSubschemas(added, revisionRefs, baseRefs)
}

// filterValidationEquivalentSubschemas keeps entries from subschemas whose
// origin-side ref has no validation-equivalent inline/$ref peer on the other
// side. originRefs is the side the entries came from (base for deleted,
// revision for added); peerRefs is the side searched for a peer.
func filterValidationEquivalentSubschemas(subschemas diff.Subschemas, originRefs, peerRefs openapi3.SchemaRefs) diff.Subschemas {
	result := diff.Subschemas{}
	for _, subschema := range subschemas {
		if subschema.Index < 0 || subschema.Index >= len(originRefs) {
			result = append(result, subschema)
			continue
		}

		if hasValidationEquivalentSubschema(originRefs[subschema.Index], peerRefs) {
			continue
		}

		result = append(result, subschema)
	}

	return result
}

func hasValidationEquivalentSubschema(schemaRef *openapi3.SchemaRef, schemaRefs openapi3.SchemaRefs) bool {
	// This is O(source x candidate); x-of lists are typically short, and the
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
