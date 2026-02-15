/*
Package allof merges allOf schema compositions into single unified schemas.

# Overview

The allof package flattens allOf compositions by merging all subschemas into one.
This improves breaking change detection accuracy because changes to individual
allOf members can be properly compared as property-level changes rather than
as entire schema replacements.

# Usage

Merge allOf in a spec:

	mergedSpec, err := allof.MergeSpec(spec)

Or use via the load package option:

	specInfo, err := load.NewSpecInfo(loader, source, load.WithFlattenAllOf())

# Merge Rules

The merge process combines schema properties following these rules:
  - Properties from all subschemas are combined
  - Required fields are merged (union)
  - Numeric constraints use the most restrictive values (e.g., max of minimums)
  - Type and format must be identical across subschemas or an error is returned
  - Enum values are intersected
  - Nested allOf compositions are recursively merged

# Example

Before:

	allOf:
	  - type: object
	    properties:
	      name: { type: string }
	  - type: object
	    properties:
	      age: { type: integer }

After:

	type: object
	properties:
	  name: { type: string }
	  age: { type: integer }
*/
package allof
