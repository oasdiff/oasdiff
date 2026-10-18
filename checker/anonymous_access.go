package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// anonymousAccess records whether an operation accepts unauthenticated
// requests before and after the change.
type anonymousAccess struct {
	known    bool
	base     bool
	revision bool
}

func declaredAnonymousAccess(operationItem *diff.MethodDiff) anonymousAccess {
	return anonymousAccess{
		// The effective security of an operation that inherits the root list
		// is unknown here, since only the operation's own list is available.
		known:    operationItem.Base.Security != nil && operationItem.Revision.Security != nil,
		base:     allowsAnonymous(operationItem.Base.Security),
		revision: allowsAnonymous(operationItem.Revision.Security),
	}
}

func effectiveAnonymousAccess(base, revision *openapi3.SecurityRequirements) anonymousAccess {
	return anonymousAccess{
		known:    true,
		base:     allowsAnonymous(base),
		revision: allowsAnonymous(revision),
	}
}

func (access anonymousAccess) changed() bool {
	return access.known && access.base != access.revision
}

// Once the revision accepts unauthenticated requests it accepts every request,
// so nothing removed from its requirements can reject a client.
func (access anonymousAccess) revisionAcceptsAll() bool {
	return access.known && access.revision
}

// allowsAnonymous reports whether a security list accepts a request without
// credentials: an empty list requires nothing, and an empty alternative (`{}`)
// is satisfied by any request.
func allowsAnonymous(securityRequirements *openapi3.SecurityRequirements) bool {
	if securityRequirements == nil || len(*securityRequirements) == 0 {
		return true
	}
	for _, securityRequirement := range *securityRequirements {
		if len(securityRequirement) == 0 {
			return true
		}
	}
	return false
}
