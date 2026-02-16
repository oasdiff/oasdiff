/*
Package commonparams moves path-level parameters to individual operations.

# Overview

OpenAPI allows defining "common parameters" at the path level that apply to all
operations under that path. This package moves those parameters into each operation,
which improves diff accuracy by ensuring parameter comparisons happen at the
operation level.

See: https://swagger.io/docs/specification/describing-parameters/

# Usage

Move common parameters in a spec:

	commonparams.Move(spec)

Or use via the load package option:

	specInfo, err := load.NewSpecInfo(loader, source, load.WithFlattenParams())

# Behavior

  - Parameters defined at path level are copied to each operation under that path
  - If an operation already has a parameter with the same name and location,
    the operation-level parameter takes precedence (no duplication)
  - After processing, path-level parameters are removed

# Example

Before:

	/users/{id}:
	  parameters:
	    - name: id
	      in: path
	  get:
	    summary: Get user

After:

	/users/{id}:
	  get:
	    summary: Get user
	    parameters:
	      - name: id
	        in: path
*/
package commonparams
