package load

import (
	"net/url"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

// refPolicy says how a recording reader applies the loader's external-refs
// policy: either by comparing against the one document the load was asked for,
// or not at all because the reader it wraps already applies it.
//
// kin-openapi applies that policy itself only when no custom ReadFromURIFunc is
// installed; with one it hands the policy to the reader (see
// Loader.loadSingleElementFromURI). A custom reader that does not apply it
// therefore turns --allow-external-refs=false into a no-op, which is how
// GHSA-2jcc-mxv7-p3f9 happened on the git-revision path and how it happened
// again in the capture reader.
//
// The two cases are separate fields rather than one path that means "no policy"
// when empty, so a caller has to name the one it wants. A zero refPolicy
// enforces and matches no document, so a caller that sets neither field refuses
// every external $ref rather than allowing every one.
type refPolicy struct {
	// root is the only document that may be read. Compared with sameLocation.
	root string

	// innerEnforces skips the comparison because the wrapped reader applies
	// the policy itself. Only the git-revision reader does: an allowed $ref
	// there keeps its "<rev>:" prefix rather than naming a file, so there is
	// no root path to compare against.
	innerEnforces bool
}

// checkLocation returns an ExternalRefError when the location may not be read,
// nil when it may, so the caller returns the error unchanged.
//
// With external refs disallowed a spec must be self-contained, so only the
// document the load was asked for may be read. That matches what kin-openapi
// refuses on its own: any $ref to another document, a sibling file included.
func (p refPolicy) checkLocation(loader *openapi3.Loader, location *url.URL) error {
	if loader == nil || loader.IsExternalRefsAllowed || p.innerEnforces {
		return nil
	}
	if sameLocation(location, p.root) {
		return nil
	}
	return &ExternalRefError{Ref: location.String()}
}

// sameLocation reports whether location is the root document. kin-openapi may
// hand the reader an absolute or a cleaned form of the path it was given, so
// compare both as given and resolved.
func sameLocation(location *url.URL, root string) bool {
	if location.String() == root {
		return true
	}
	lp := filepath.FromSlash(location.Path)
	if lp == root || filepath.Clean(lp) == filepath.Clean(root) {
		return true
	}
	a, errA := filepath.Abs(lp)
	b, errB := filepath.Abs(root)
	return errA == nil && errB == nil && a == b
}
