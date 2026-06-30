/*
Package review builds the encrypted review bundle: the two specs, the computed
changelog, and the per-change structural blocks a review UI renders as cards.

It is the single source of truth for the bundle's on-the-wire shape (Payload),
its encryption (Encrypt), and the per-change fingerprint manifest (Manifest); a
decryptor on the rendering side mirrors the same layout.

The bundle is zero-knowledge by construction: Encrypt seals the Payload with a
fresh AES-256-GCM key and returns the ciphertext and the key separately. The
caller uploads only the ciphertext and keeps the key out of band, so the server
receives a blob it cannot read. This package makes no assumption about where the
bundle is uploaded or rendered; that is the caller's concern.

# Blocks: render only what changed

A side-by-side review that renders both specs in full is unusable for a large
spec (tens to hundreds of thousands of lines commit ~1M DOM nodes). Extract
slices the spec into just the structural blocks that changed, so the UI renders
one self-contained card per change while keeping each change in context:

	blocks := review.Extract(changes, baseSpec, revisionSpec, baseText, revText)

Each Block carries its key/title (e.g. "POST /users" or
"components/schemas/User"), the ids and fingerprints of the changes inside it,
and the block's source-text slice on each side with its starting line.

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
Known gaps: top-level sections (info/servers/tags/security) are not indexed yet;
multi-file specs (a $ref'd block defined in another file) are not sliced yet;
and because blocks are keyed off the changelog, a block whose only change has no
changelog entry (e.g. a description-only edit) is not shown, that schema-shape
completeness is a later phase. Within a shown block the slice is the raw text
diff, so unflagged changed lines are still visible.
*/
package review
