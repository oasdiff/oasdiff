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

func declaredAnonymousAccess(methodDiff *diff.MethodDiff) anonymousAccess {
	return anonymousAccess{
		// An operation without its own list inherits the root list, which is
		// not available here.
		known:    methodDiff.Base.Security != nil && methodDiff.Revision.Security != nil,
		base:     allowsAnonymous(methodDiff.Base.Security),
		revision: allowsAnonymous(methodDiff.Revision.Security),
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

// securityAlternatives lists each alternative the way the other security
// findings name one.
func securityAlternatives(securityRequirements openapi3.SecurityRequirements) []string {
	alternatives := make([]string, len(securityRequirements))
	for i, securityRequirement := range securityRequirements {
		alternatives[i] = diff.SecurityAlternative{Schemes: securityRequirement}.String()
	}
	return alternatives
}
