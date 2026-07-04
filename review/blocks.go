package review

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// fileBase is the display filename for a block's source file: the basename,
// with any git "<rev>:" prefix stripped.
func fileBase(f string) string {
	if f == "" {
		return ""
	}
	b := filepath.Base(f)
	if i := strings.LastIndex(b, ":"); i >= 0 {
		b = b[i+1:] // strip a git "<rev>:" prefix that survived Base (no "/" in the ref)
	}
	return b
}

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
	BaseFile      string `json:"base_file,omitempty"` // basename of the base slice's source file (a $ref'd file differs from the root)
	BaseText      string `json:"base_text"`           // source slice on the base side ("" if absent)
	BaseLineStart int    `json:"base_line_start"`     // 1-based first line of BaseText in the base spec
	RevFile       string `json:"rev_file,omitempty"`  // basename of the revision slice's source file
	RevText       string `json:"rev_text"`            // source slice on the revision side ("" if absent)
	RevLineStart  int    `json:"rev_line_start"`      // 1-based first line of RevText in the revision spec
}

// Extract groups changes by their enclosing structural block and slices each
// block's text, ordered by first appearance. The docs must be loaded with
// IncludeOrigin (end positions). baseTexts/revTexts map each contributing
// file's path, as reported on element origins, to its raw source (see
// load.NewSpecInfoWithCapture), so a block slices from the file it lives in.
func Extract(changes checker.Changes, baseDocs, revDocs []*openapi3.T, baseTexts, revTexts map[string]string) []Block {
	baseIdx := buildIndex(baseDocs...)
	revIdx := buildIndex(revDocs...)

	byKey := map[string]*Block{}
	spansByKey := map[string]resolution{}
	var order []string
	for _, c := range changes {
		r := resolve(c, baseIdx, revIdx)
		b := byKey[r.key]
		if b == nil {
			b = &Block{Key: r.key, Title: r.title}
			byKey[r.key] = b
			spansByKey[r.key] = r
			order = append(order, r.key)
		}
		b.ChangeIDs = append(b.ChangeIDs, c.GetId())
		b.Fingerprints = append(b.Fingerprints, formatters.ComputeFingerprint(c.GetId(), c.GetOperation(), c.GetPath(), c.GetArgs()))
	}

	out := make([]Block, 0, len(order))
	for _, key := range order {
		b := byKey[key]
		r := spansByKey[key]
		if r.base != nil {
			b.BaseFile = fileBase(r.base.file)
			b.BaseText, b.BaseLineStart = sliceLines(textFor(baseTexts, r.base.file), r.base.start, r.base.end), r.base.start
		}
		if r.rev != nil {
			b.RevFile = fileBase(r.rev.file)
			b.RevText, b.RevLineStart = sliceLines(textFor(revTexts, r.rev.file), r.rev.start, r.rev.end), r.rev.start
		}
		out = append(out, *b)
	}
	return out
}

// textFor looks up a span's origin file in the captured texts. Capture keys
// have no leading "./"; origins may carry one (net/url prepends it when a
// relative path's first segment contains a colon, as git refs do).
func textFor(texts map[string]string, file string) string {
	if t, ok := texts[file]; ok {
		return t
	}
	return texts[strings.TrimPrefix(file, "./")]
}

// resolution is a change's block assignment: its card key/title and the exact
// span to slice on each side (nil = no slice on that side).
type resolution struct {
	key, title string
	base, rev  *span
}

// resolve decides which block a change belongs to: the smallest block
// containing its source line (revision side first, then base), which follows
// $refs. The matched span is carried through to slicing so a key existing in
// several files (composed) slices the file the change is in. Fallback when no
// source line resolves: (operation, path), then the rule Area, then "Other".
func resolve(c checker.Change, baseIdx, revIdx docIndex) resolution {
	baseSrc, revSrc := changeSources(c)
	revSpan := matchSpan(revIdx, revSrc)
	baseSpan := matchSpan(baseIdx, baseSrc)
	switch {
	case revSpan != nil:
		if baseSpan == nil || baseSpan.key != revSpan.key {
			baseSpan = baseIdx.spanFor(revSpan.key, revSpan.file)
		}
		return newResolution(revSpan.key, baseSpan, revSpan, baseIdx, revIdx)
	case baseSpan != nil:
		return newResolution(baseSpan.key, baseSpan, revIdx.spanFor(baseSpan.key, baseSpan.file), baseIdx, revIdx)
	}
	key, title := fallbackKey(c)
	return resolution{key: key, title: title, base: baseIdx.spanFor(key, ""), rev: revIdx.spanFor(key, "")}
}

func matchSpan(idx docIndex, src *checker.Source) *span {
	if src == nil || src.Line <= 0 {
		return nil
	}
	if s, ok := idx.smallestContaining(src.File, src.Line); ok {
		return &s
	}
	return nil
}

// newResolution qualifies the card key with its filename when the same key
// exists in more than one file (composed), keeping each file's block separate.
func newResolution(key string, base, rev *span, baseIdx, revIdx docIndex) resolution {
	display := key
	if len(baseIdx.byKey[key]) > 1 || len(revIdx.byKey[key]) > 1 {
		if f := firstFileBase(rev, base); f != "" {
			display = f + ": " + key
		}
	}
	return resolution{key: display, title: display, base: base, rev: rev}
}

func firstFileBase(spans ...*span) string {
	for _, s := range spans {
		if s != nil {
			if f := fileBase(s.file); f != "" {
				return f
			}
		}
	}
	return ""
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

// changeSources reads via the Change interface so SecurityChange and
// ComponentChange resolve to their blocks too, not only ApiChange.
func changeSources(c checker.Change) (base, rev *checker.Source) {
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
// the smallest block containing a line, and the spans for a key. A key can map
// to several spans when composed specs define the same component name or
// top-level section in more than one file.
type docIndex struct {
	spans []span
	byKey map[string][]span
}

// buildIndex enumerates the sliceable structural blocks of one or more docs
// (composed mode shares one index): path items, operations, and named
// components, all of which carry Origin. Component names and top-level
// sections repeat across composed specs, so byKey keeps every span and
// resolution disambiguates by file.
func buildIndex(docs ...*openapi3.T) docIndex {
	idx := docIndex{byKey: map[string][]span{}}
	add := func(key, title string, o *openapi3.Origin) {
		if start, end, ok := originEndRange(o); ok {
			s := span{key: key, title: title, file: o.Key.File, start: start, end: end}
			idx.spans = append(idx.spans, s)
			idx.byKey[key] = append(idx.byKey[key], s)
		}
	}
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		addDoc(&idx, doc, add)
	}
	return idx
}

// spanFor returns the span to slice for key: the unique one, or, among several
// (composed specs repeating a name), the one from preferredFile. Nil when
// absent or ambiguous: a missing slice is safe, the wrong file's slice is not.
func (idx docIndex) spanFor(key, preferredFile string) *span {
	entries := idx.byKey[key]
	if len(entries) == 1 {
		return &entries[0]
	}
	if preferredFile != "" {
		for i := range entries {
			if fileBase(entries[i].file) == fileBase(preferredFile) {
				return &entries[i]
			}
		}
	}
	return nil
}

// addDoc indexes one document's blocks into idx.
func addDoc(idx *docIndex, doc *openapi3.T, add func(key, title string, o *openapi3.Origin)) {
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
		for name, ref := range doc.Components.SecuritySchemes {
			if ref != nil && ref.Value != nil {
				k := "components/securitySchemes/" + name
				add(k, k, ref.Value.Origin)
			}
		}
	}
	addTopLevelSections(idx, doc)
	indexExternalSchemas(idx, doc)
}

// indexExternalSchemas indexes schemas $ref'd from another file: they live at
// the usage site (not in this doc's Components) and carry the external file's
// origin. Keyed by the ref, stable across sides and deduped across operations.
func indexExternalSchemas(idx *docIndex, doc *openapi3.T) {
	_ = doc.WalkSchemas(func(_ string, sr *openapi3.SchemaRef) error {
		if sr == nil || sr.Value == nil {
			return nil
		}
		// a non-empty part before "#" = another file; internal "#/..." refs are already indexed
		if before, _, _ := strings.Cut(sr.Ref, "#"); before == "" {
			return nil
		}
		if start, end, ok := originEndRange(sr.Value.Origin); ok {
			key := strings.TrimPrefix(sr.Ref, "./")
			if len(idx.byKey[key]) > 0 {
				return nil
			}
			s := span{key: key, title: key, file: sr.Value.Origin.Key.File, start: start, end: end}
			idx.spans = append(idx.spans, s)
			idx.byKey[key] = append(idx.byKey[key], s)
		}
		return nil
	})
}

// addTopLevelSections indexes info/servers/tags/security, each spanning from
// its key line to just before the next top-level key. Key-line based on
// purpose: security items are bare maps with no origin of their own, so
// per-element spans aren't available.
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
				return l - 1
			}
		}
		return math.MaxInt // last section: to EOF (sliceLines clamps)
	}
	for _, name := range []string{"info", "servers", "tags", "security"} {
		if loc, ok := doc.Origin.Fields[name]; ok && loc.Line > 0 {
			s := span{key: name, title: name, file: loc.File, start: loc.Line, end: sectionEnd(loc.Line)}
			idx.spans = append(idx.spans, s)
			idx.byKey[name] = append(idx.byKey[name], s)
		}
	}
}

// smallestContaining returns the most specific block in file containing line.
// Matching on file keeps multi-file specs correct: line numbers are not unique
// across files.
func (idx docIndex) smallestContaining(file string, line int) (span, bool) {
	best := span{}
	bestSize := math.MaxInt
	found := false
	for _, s := range idx.spans {
		if fileMatches(s.file, file) && line >= s.start && line <= s.end {
			if size := s.end - s.start; size < bestSize {
				best, bestSize, found = s, size, true
			}
		}
	}
	return best, found
}

// fileMatches accepts both forms of the same file: origin files keep a git
// "<rev>:" prefix, checker sources are display paths with it stripped.
func fileMatches(spanFile, srcFile string) bool {
	return spanFile == srcFile || strings.HasSuffix(spanFile, ":"+srcFile)
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
// to its bounds; "" for an invalid span.
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
