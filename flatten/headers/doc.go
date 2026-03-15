/*
Package headers normalizes HTTP header names to lowercase.

# Overview

HTTP headers are case-insensitive per RFC 7230. This package converts all header
parameter names and response header names to lowercase, ensuring consistent
comparison during diff operations regardless of how headers were originally cased.

# Usage

Lowercase headers in a spec:

	headers.Lowercase(spec)

Or use via the load package option:

	specInfo, err := load.NewSpecInfo(loader, source, load.WithLowercaseHeaders())

# Scope

The package processes:
  - Header parameters at path level
  - Header parameters at operation level
  - Response headers in all responses

# Example

Before:

	parameters:
	  - name: X-Request-ID
	    in: header
	  - name: Authorization
	    in: header

After:

	parameters:
	  - name: x-request-id
	    in: header
	  - name: authorization
	    in: header
*/
package headers
