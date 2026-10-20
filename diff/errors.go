package diff

import "fmt"

// DuplicateEndpointError reports two operations that are the same endpoint
// and have the same x-since-date, so neither can be chosen.
type DuplicateEndpointError struct {
	Method  string
	Path1   string
	Source1 string
	Path2   string
	Source2 string
}

func (e *DuplicateEndpointError) Error() string {
	hint := fmt.Sprintf("add the %s extension to specify order", SinceDateExtension)
	// different paths only match because path parameter names are ignored
	if e.Path1 != e.Path2 {
		hint += ", or include path parameter names in endpoint matching with --include-path-params"
	}
	return fmt.Sprintf("duplicate endpoint: %s %s in %s and %s %s in %s. You may %s",
		e.Method, e.Path1, e.Source1, e.Method, e.Path2, e.Source2, hint)
}

func invalidSinceDateError(source, op, path string, err error) error {
	return fmt.Errorf("invalid %s extension value in %s (%s %s), %w", SinceDateExtension, source, op, path, err)
}
