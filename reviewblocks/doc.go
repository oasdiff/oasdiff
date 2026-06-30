/*
Package reviewblocks slices an OpenAPI spec into the structural blocks that
changed, so a review can render one self-contained card per change instead of
the whole document.

# Why

The side-by-side review renders both specs in full. For a large spec (tens to
hundreds of thousands of lines) that commits ~1M DOM nodes and is slow to become
interactive. Rendering only the changed blocks drops that by ~99% while keeping
each change in its surrounding context.

# What it produces

Extract groups a checker.Changes list by the structural block each change falls
in, and returns one Block per group:

	blocks := reviewblocks.Extract(changes, baseSpec, revisionSpec, baseText, revText)

Each Block carries its key/title (e.g. "POST /users" or
"components/schemas/User"), the ids and fingerprints of the changes inside it,
and the block's source-text slice on each side with its starting line. The
review page renders the slice as a side-by-side diff and overlays the changes
as cards; the fingerprints are the stable key it joins each change to its block.

# Block selection

A change is keyed to the smallest indexed block whose origin span contains its
source line, not by its (operation, path). This matters for $refs: a change
inside a $ref'd component is reported under the referencing operation, but its
source line is in the component, so keying by line follows the $ref and cards it
as the component, and dedupes the same component change reported across several
operations into one card. When no source line resolves (e.g. a change detected
after --flatten-allof, whose merged schema has no single location), it falls
back to the operation it names, then the rule Area, then an "other changes"
bucket.

# Slicing

Slicing relies on kin-openapi origin end positions
(openapi3.Origin.Key.EndLine/EndColumn) so a block's full extent is known, not
just its start. The specs must be loaded with IncludeOrigin = true.

# Status and limitations

Indexed block types are operations, path items, and named component schemas.
Known gaps, tracked in the review-page design: top-level sections
(info/servers/tags/security) are not indexed yet; multi-file specs (a $ref'd
block defined in another file) are not sliced yet; and because blocks are keyed
off the changelog, a block whose only change has no changelog entry (e.g. a
description-only edit) is not shown, that schema-shape completeness is a later
phase. Within a shown block the slice is the raw text diff, so unflagged changed
lines are still visible.
*/
package reviewblocks
