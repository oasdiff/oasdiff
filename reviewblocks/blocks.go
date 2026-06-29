// Package reviewblocks extracts, per reported change, the smallest enclosing
// OpenAPI structural block (operation / path / named component / top-level) and
// its source-text slice, so the review page can render one self-contained card
// per change instead of the full spec. This is the large-spec rendering fix:
// the full-spec side-by-side commits ~1M DOM nodes for a 250k-line spec; cards
// drop that by ~99%.
//
// Which block a change belongs to is decided by the change's source line, not
// its (operation, path): a change inside a $ref'd component is reported under
// the referencing operation, but its source line is in the component, so
// keying by the line (matched against block origin spans) follows the $ref and
// cards it as the component (and dedupes the same component change reported
// across several endpoints into one card). The (operation, path) and the rule
// Area are the fallback when no source line resolves.
//
// Slicing relies on kin-openapi origin end-positions
// (openapi3.Origin.Key.EndLine/EndColumn) so a block's full extent is known,
// not just its start. The specs must be loaded with IncludeOrigin = true.
//
// WIP / draft. Wired: change grouping by source-line containment, the
// Origin-end slice, and the operation / path / named-component-schema blocks.
// Still to do (see the design): other named components, top-level blocks
// (info/servers/tags/security), overlap dedup, and feeding the blocks into the
// encrypted review bundle so the browser renders cards from decrypted blocks
// (the server never sees the spec).
package reviewblocks

import (
	"math"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// otherChangesKey collects changes with no resolvable block; they render as a
// single "Other changes" card with no side-by-side.
const otherChangesKey = "__other__"

// Block is one review card: the source-text slice of a structural block on each
// side, plus the ids of the changes that fall inside it. Empty BaseText/RevText
// means that side has no sliceable source (e.g. an added or removed block, or a
// location that did not resolve to a block).
type Block struct {
	Key          string   `json:"key"`          // stable identity, e.g. "POST /users" or "components/schemas/User"
	Title        string   `json:"title"`        // human header
	ChangeIDs    []string `json:"change_ids"`   // rule ids of the changes in this block (for display/debug)
	Fingerprints []string `json:"fingerprints"` // per-change fingerprints, aligned with ChangeIDs; the
	// stable key the review page joins each change to its card on
	BaseText      string `json:"base_text"`       // source slice on the base side ("" if absent)
	BaseLineStart int    `json:"base_line_start"` // 1-based first line of BaseText in the base spec
	RevText       string `json:"rev_text"`        // source slice on the revision side ("" if absent)
	RevLineStart  int    `json:"rev_line_start"`  // 1-based first line of RevText in the revision spec
}

// Extract groups changes by their enclosing structural block and slices each
// block's text from the base and revision specs. Order follows first
// appearance in changes. The specs must be loaded with IncludeOrigin so
// end-positions are available; baseText/revText are the raw spec sources.
func Extract(changes checker.Changes, base, revision *openapi3.T, baseText, revText string) []Block {
	baseIdx := buildIndex(base)
	revIdx := buildIndex(revision)

	byKey := map[string]*Block{}
	var order []string
	for _, c := range changes {
		key, title := blockKey(c, baseIdx, revIdx)
		b := byKey[key]
		if b == nil {
			b = &Block{Key: key, Title: title}
			byKey[key] = b
			order = append(order, key)
		}
		b.ChangeIDs = append(b.ChangeIDs, c.GetId())
		b.Fingerprints = append(b.Fingerprints, formatters.ComputeFingerprint(c.GetId(), c.GetOperation(), c.GetPath(), c.GetArgs()))
	}

	out := make([]Block, 0, len(order))
	for _, key := range order {
		b := byKey[key]
		if s, ok := baseIdx.byKey[key]; ok {
			b.BaseText, b.BaseLineStart = sliceLines(baseText, s.start, s.end), s.start
		}
		if s, ok := revIdx.byKey[key]; ok {
			b.RevText, b.RevLineStart = sliceLines(revText, s.start, s.end), s.start
		}
		out = append(out, *b)
	}
	return out
}

// blockKey decides which structural block a change belongs to. Primary signal:
// the change's source line matched against the smallest enclosing block, on the
// revision side (current state) first, then the base side (deletions). This
// follows $refs. Fallback when no source line resolves: the (operation, path)
// key, then the rule Area (top-level changes), then "Other changes".
func blockKey(c checker.Change, baseIdx, revIdx docIndex) (key, title string) {
	baseSrc, revSrc := changeSources(c)
	if revSrc != nil && revSrc.Line > 0 {
		if s, ok := revIdx.smallestContaining(revSrc.Line); ok {
			return s.key, s.title
		}
	}
	if baseSrc != nil && baseSrc.Line > 0 {
		if s, ok := baseIdx.smallestContaining(baseSrc.Line); ok {
			return s.key, s.title
		}
	}
	return fallbackKey(c)
}

func fallbackKey(c checker.Change) (key, title string) {
	switch op, path := c.GetOperation(), c.GetPath(); {
	case op != "" && path != "":
		k := op + " " + path
		return k, k
	case path != "":
		return path, path
	}
	// No operation/path: bucket top-level changes by their Area name (security,
	// tags, ...). Schema changes with no path fall through to "Other" rather
	// than being mis-bucketed.
	if area, ok := areaByID[c.GetId()]; ok && area != checker.AreaSchema {
		a := area.String()
		return a, a
	}
	return otherChangesKey, "Other changes"
}

// changeSources returns the change's base and revision source locations, which
// oasdiff resolves to the changed element (following $refs). Nil when absent or
// when the change is not the concrete ApiChange.
func changeSources(c checker.Change) (base, rev *checker.Source) {
	if ac, ok := c.(checker.ApiChange); ok {
		return ac.BaseSource, ac.RevisionSource
	}
	return nil, nil
}

// areaByID maps each rule id to its Area, so a change's Area can be looked up
// from its id for the fallback bucketing.
var areaByID = func() map[string]checker.Area {
	rules := checker.GetAllRules()
	m := make(map[string]checker.Area, len(rules))
	for _, r := range rules {
		m[r.Id] = r.Area
	}
	return m
}()

// span is one enumerated structural block in a single document: its key/title
// and 1-based inclusive line span.
type span struct {
	key, title string
	start, end int
}

// docIndex indexes a document's enumerated structural blocks for two lookups:
// the smallest block containing a line, and the span for a key.
type docIndex struct {
	spans []span
	byKey map[string]span
}

// buildIndex enumerates the sliceable structural blocks of doc: every path
// item, every operation, and every named component schema (all carry Origin).
func buildIndex(doc *openapi3.T) docIndex {
	idx := docIndex{byKey: map[string]span{}}
	add := func(key, title string, o *openapi3.Origin) {
		if start, end, ok := originEndRange(o); ok {
			s := span{key: key, title: title, start: start, end: end}
			idx.spans = append(idx.spans, s)
			idx.byKey[key] = s
		}
	}
	if doc == nil {
		return idx
	}
	if doc.Paths != nil {
		for path, pi := range doc.Paths.Map() {
			if pi == nil {
				continue
			}
			add(path, path, pi.Origin)
			for method, op := range pi.Operations() {
				k := method + " " + path
				add(k, k, op.Origin)
			}
		}
	}
	if doc.Components != nil {
		for name, ref := range doc.Components.Schemas {
			if ref != nil && ref.Value != nil {
				k := "components/schemas/" + name
				add(k, k, ref.Value.Origin)
			}
		}
	}
	return idx
}

// smallestContaining returns the span containing line whose extent is smallest
// (the most specific enclosing block), and whether one was found.
func (idx docIndex) smallestContaining(line int) (span, bool) {
	best := span{}
	bestSize := math.MaxInt
	found := false
	for _, s := range idx.spans {
		if line >= s.start && line <= s.end {
			if size := s.end - s.start; size < bestSize {
				best, bestSize, found = s, size, true
			}
		}
	}
	return best, found
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
