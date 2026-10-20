package diff

import "fmt"

// DuplicateEndpointError reports two operations that are the same endpoint
// and have the same x-since-date, so neither can be chosen. Returned as a
// distinct type so callers can use errors.As to tell how the two paths match.
type DuplicateEndpointError struct {
	Method  string
	Path1   string
	Source1 string
	Path2   string
	Source2 string
}

func (e *DuplicateEndpointError) Error() string {
	return fmt.Sprintf("duplicate endpoint: %s %s in %s and %s %s in %s. You may add the %s extension to specify order",
		e.Method, e.Path1, e.Source1, e.Method, e.Path2, e.Source2, SinceDateExtension)
}
