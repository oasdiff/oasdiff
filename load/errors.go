package load

import "fmt"

// FlattenError reports a failure to merge allOf during WithFlattenAllOf.
// Returned wrapped so callers can use errors.As to distinguish a flatten
// failure (which happens after the spec has loaded successfully) from a
// genuine load failure.
type FlattenError struct {
	Url string
	Err error
}

func (e *FlattenError) Error() string {
	return fmt.Sprintf("failed to flatten allOf in %q: %s", e.Url, e.Err)
}

func (e *FlattenError) Unwrap() error { return e.Err }

// ExternalRefError reports that a spec resolved an external $ref (an http(s)
// URL or a local file outside the git tree) while external refs were disallowed
// (IsExternalRefsAllowed=false, i.e. --allow-external-refs=false). Returned as a
// distinct type so callers can use errors.As to map it to a dedicated exit code,
// rather than matching the message text.
type ExternalRefError struct {
	Ref string
}

func (e *ExternalRefError) Error() string {
	return fmt.Sprintf("external $ref not allowed (enable --allow-external-refs to permit): %s", e.Ref)
}
