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
