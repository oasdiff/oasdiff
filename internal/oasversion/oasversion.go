// Package oasversion tells which reading of a document its OpenAPI version
// calls for, where OpenAPI 3.0 and OpenAPI 3.1 read the same text differently.
package oasversion

import "github.com/getkin/kin-openapi/openapi3"

// Dialect is the reading a document's OpenAPI version gives it. OpenAPI 3.1
// schemas are JSON Schema 2020-12, while OpenAPI 3.0 schemas follow the
// specification's own subset, and a few keywords mean different things under
// each.
type Dialect int

const (
	OpenAPI30 Dialect = iota
	OpenAPI31
)

// Of returns the dialect of a document from its openapi version. OpenAPI 3.1
// and every later version read as OpenAPI31, and anything else, an
// unrecognized version included, as OpenAPI30.
func Of(version string) Dialect {
	if (&openapi3.T{OpenAPI: version}).IsOpenAPI31OrLater() {
		return OpenAPI31
	}
	return OpenAPI30
}

// Dialects are the dialects of the two documents being compared. They differ
// when the revision moves to another OpenAPI version.
type Dialects struct {
	Base, Revision Dialect
}
