package review

import (
	"math"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// otherChangesKey collects changes with no resolvable block; they group under a
// single "Other changes" key with no source slice.
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
// block's text from the base and revision specs. Order follows first appearance
// in changes. The specs must be loaded with IncludeOrigin so end-positions are
// available. baseTexts/revTexts map each contributing file's path (as reported
// on element origins) to its raw source; for a single-file spec that's one
// entry. A block is sliced from the file it lives in, so a $ref'd-from-another-
// file block is sliced from that file (see load.NewSpecInfoWithCapture).
func Extract(changes checker.Changes, base, revision *openapi3.T, baseTexts, revTexts map[string]string) []Block {
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
			b.BaseText, b.BaseLineStart = sliceLines(baseTexts[s.file], s.start, s.end), s.start
		}
		if s, ok := revIdx.byKey[key]; ok {
			b.RevText, b.RevLineStart = sliceLines(revTexts[s.file], s.start, s.end), s.start
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
		if s, ok := revIdx.smallestContaining(revSrc.File, revSrc.Line); ok {
			return s.key, s.title
		}
	}
	if baseSrc != nil && baseSrc.Line > 0 {
		if s, ok := baseIdx.smallestContaining(baseSrc.File, baseSrc.Line); ok {
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
	// The interface carries the sources for every change type (ApiChange,
	// ComponentChange, SecurityChange all embed CommonChange), so a security or
	// component change resolves to its block, not just operation changes.
	return c.GetBaseSource(), c.GetRevisionSource()
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

// span is one enumerated structural block: its key/title, the file it lives in
// (the element's origin File; "" for an in-memory spec), and its 1-based
// inclusive line range within that file.
type span struct {
	key, title string
	file       string
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
			s := span{key: key, title: title, file: o.Key.File, start: start, end: end}
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
	addTopLevelSections(&idx, doc)
	indexExternalSchemas(&idx, doc)
	return idx
}

// indexExternalSchemas indexes schemas $ref'd from another file. They live at
// the usage site (not in this doc's Components) and carry the external file's
// origin, so a change inside one is keyed to that block and sliced from that
// file. Keyed by the ref (stable across base/revision and deduped across the N
// operations that share it).
func indexExternalSchemas(idx *docIndex, doc *openapi3.T) {
	_ = doc.WalkSchemas(func(_ string, sr *openapi3.SchemaRef) error {
		if sr == nil || sr.Value == nil {
			return nil
		}
		// A non-empty part before "#" means the ref points at another file
		// (internal "#/..." refs have none and are already indexed).
		if before, _, _ := strings.Cut(sr.Ref, "#"); before == "" {
			return nil
		}
		if start, end, ok := originEndRange(sr.Value.Origin); ok {
			key := strings.TrimPrefix(sr.Ref, "./")
			if _, exists := idx.byKey[key]; exists {
				return nil
			}
			s := span{key: key, title: key, file: sr.Value.Origin.Key.File, start: start, end: end}
			idx.spans = append(idx.spans, s)
			idx.byKey[key] = s
		}
		return nil
	})
}

// addTopLevelSections indexes the document-level sections (info / servers /
// tags / security). Each spans from its key line to just before the next
// top-level key, derived from the document origin's field locations. This is
// key-line based on purpose: it works uniformly even for `security`, whose
// items are bare maps (SecurityRequirement) with no origin of their own, where
// per-element spans (as used for operations and component schemas) aren't
// available.
func addTopLevelSections(idx *docIndex, doc *openapi3.T) {
	if doc.Origin == nil {
		return
	}
	keyLines := make([]int, 0, len(doc.Origin.Fields))
	for _, loc := range doc.Origin.Fields {
		if loc.Line > 0 {
			keyLines = append(keyLines, loc.Line)
		}
	}
	sort.Ints(keyLines)
	sectionEnd := func(start int) int {
		for _, l := range keyLines {
			if l > start {
				return l - 1 // up to just before the next top-level key
			}
		}
		return math.MaxInt // last section: to EOF (sliceLines clamps)
	}
	for _, name := range []string{"info", "servers", "tags", "security"} {
		if loc, ok := doc.Origin.Fields[name]; ok && loc.Line > 0 {
			s := span{key: name, title: name, file: loc.File, start: loc.Line, end: sectionEnd(loc.Line)}
			idx.spans = append(idx.spans, s)
			idx.byKey[name] = s
		}
	}
}

// smallestContaining returns the span in file containing line whose extent is
// smallest (the most specific enclosing block), and whether one was found.
// Matching on file is what keeps multi-file specs correct: line numbers are not
// unique across files.
func (idx docIndex) smallestContaining(file string, line int) (span, bool) {
	best := span{}
	bestSize := math.MaxInt
	found := false
	for _, s := range idx.spans {
		if s.file == file && line >= s.start && line <= s.end {
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
