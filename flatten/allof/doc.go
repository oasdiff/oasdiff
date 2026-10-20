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

# Recursive schemas

A document can only express recursion through a $ref, so the merged output
maintains one invariant: every cycle-closing edge carries a $ref. A cycle
through a named component keeps its name. A cycle through the anonymous
result of merging recursive branches is hoisted into components.schemas
under a name built from the merged component names
(AllOfMerged_NodeA_NodeB), so the same cycle keeps its name in every
revision and flatten produces the same output standalone and inside diff. A node that combines such a
cycle with further constraints keeps them as a residual allOf of the named
cycle and the merged rest: complete, just not flattened at that one node.

MergeSpec maintains the invariant for a whole document. Merge flattens a
single schema without access to a components section, so a recursive input
whose cycle has no name merges to an in-memory cycle with no $ref, which
does not marshal.

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
