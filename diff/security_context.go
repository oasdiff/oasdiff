package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// SecurityContext carries what a checker needs to work out the security that
// applies to each operation. An operation without its own security list
// inherits the document-root one, so a change to the root list changes
// operations whose own diff is empty, and those operations appear nowhere else
// in the diff.
type SecurityContext struct {
	Base     openapi3.SecurityRequirements
	Revision openapi3.SecurityRequirements

	// Operations holds every operation present in both specs, matched and
	// filtered the same way as PathsDiff, whether or not it changed.
	Operations []OperationPair
}
