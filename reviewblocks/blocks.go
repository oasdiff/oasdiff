// Package reviewblocks extracts, per reported change, the smallest enclosing
// OpenAPI structural block (operation / path / named component / top-level) and
// its source-text slice, so the review page can render one self-contained card
// per change instead of the full spec. This is the large-spec rendering fix:
// the full-spec side-by-side commits ~1M DOM nodes for a 250k-line spec; cards
// drop that by ~99%.
//
// Slicing relies on kin-openapi origin end-positions
// (openapi3.Origin.Key.EndLine/EndColumn, added upstream) so a block's full
// extent is known, not just its start. The specs must be loaded with
// IncludeOrigin = true.
//
// WIP / draft. Wired so far: change grouping, the Origin-end slice, and the
// component-schema case. Still to do (see the design): operation / path-item /
// top-level node resolution (operations don't carry Origin directly, so the
// path-item key origin is the handle), overlap dedup, and the "Other changes"
// fallback rendering. Then this feeds the encrypted review bundle so the
// browser renders cards from decrypted blocks (the server never sees the spec).
package reviewblocks

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
)

// otherChangesKey collects changes whose structural location we don't (yet) map
// to a sliceable block; they render as a single "Other changes" card with no
// side-by-side.
const otherChangesKey = "__other__"

// Block is one review card: the source-text slice of a structural block on each
// side, plus the ids of the changes that fall inside it. Empty BaseText/RevText
// means that side has no sliceable source (e.g. an added or removed block, or a
// location not yet resolved).
type Block struct {
	Key           string   // stable identity, e.g. "POST /users" or "components/schemas/User"
	Title         string   // human header
	ChangeIDs     []string // changes contained in this block
	BaseText      string
	BaseLineStart int // 1-based first line of BaseText in the base spec
	RevText       string
	RevLineStart  int // 1-based first line of RevText in the revision spec
}

// Extract groups changes by their enclosing structural block and slices each
// block's text from the base and revision specs. Order follows first
// appearance in changes. The specs must be loaded with IncludeOrigin so
// end-positions are available; baseText/revText are the raw spec sources.
func Extract(changes checker.Changes, base, revision *openapi3.T, baseText, revText string) []Block {
	byKey := map[string]*Block{}
	var order []string
	for _, c := range changes {
		key, title := structuralKey(c)
		b := byKey[key]
		if b == nil {
			b = &Block{Key: key, Title: title}
			byKey[key] = b
			order = append(order, key)
		}
		b.ChangeIDs = append(b.ChangeIDs, c.GetId())
	}

	out := make([]Block, 0, len(order))
	for _, key := range order {
		b := byKey[key]
		b.BaseText, b.BaseLineStart = sliceFor(base, baseText, key)
		b.RevText, b.RevLineStart = sliceFor(revision, revText, key)
		out = append(out, *b)
	}
	return out
}

// structuralKey maps a change to the key of its smallest enclosing structural
// block. WIP: keys operations as "<METHOD> <path>", path-level and named
// components by their path string, and everything else into the "Other
// changes" fallback. The exact shape of a component change's path is a design
// open question (sample real changelogs before locking the mapping).
func structuralKey(c checker.Change) (key, title string) {
	op, path := c.GetOperation(), c.GetPath()
	switch {
	case op != "" && path != "":
		k := op + " " + path
		return k, k
	case path != "":
		return path, path
	default:
		return otherChangesKey, "Other changes"
	}
}

// sliceFor resolves a structural key to its node in spec, reads the origin
// end-range, and slices the source text. WIP: only the component-schema case is
// wired (schemas carry Origin); operation / path-item / top-level cases return
// empty until their node resolution lands.
func sliceFor(spec *openapi3.T, text, key string) (string, int) {
	if spec == nil || spec.Components == nil {
		return "", 0
	}
	if name, ok := strings.CutPrefix(key, "components/schemas/"); ok {
		if ref := spec.Components.Schemas[name]; ref != nil && ref.Value != nil {
			if start, end, ok := originEndRange(ref.Value.Origin); ok {
				return sliceLines(text, start, end), start
			}
		}
	}
	// TODO(endpoint-review): operation (path-item key origin), path-level,
	// other named components, top-level (info/servers/tags/security), and the
	// "Other changes" fallback. Dedup overlapping ranges.
	return "", 0
}

// originEndRange returns the 1-based [start, end] line span a key origin heads,
// using the kin-openapi end-position support. ok is false when origin or end
// info is absent (older yaml, or a node with no recorded end).
func originEndRange(o *openapi3.Origin) (start, end int, ok bool) {
	if o == nil || o.Key == nil || o.Key.Line == 0 || o.Key.EndLine == 0 {
		return 0, 0, false
	}
	return o.Key.Line, o.Key.EndLine, true
}

// sliceLines returns lines [start, end] (1-based, inclusive) of text, clamped
// to its bounds. Returns "" for an out-of-range or inverted span.
func sliceLines(text string, start, end int) string {
	if start < 1 || end < start {
		return ""
	}
	lines := strings.Split(text, "\n")
	if start > len(lines) {
		return ""
	}
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start-1:end], "\n")
}
