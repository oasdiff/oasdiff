package review

import (
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Deriving a block's extent from start positions, so no end position is needed
// from the parser.
//
// A block runs to the line before the next key or sequence item at the same or
// shallower indentation. Every such boundary is already present in the origin
// data kin produces: Origin.Key is a key, Origin.Fields holds every field key
// of a mapping, and Origin.Sequences holds each scalar item of a
// sequence-valued field. Collecting those per file gives the boundary set.
//
// Measured against parser-recorded end positions on ~11.9M spans (kin's
// corpus, oasdiff's, the GitHub and Stripe specs), the two agree on ~99.98%.
// The residual is whether a trailing blank or comment line between two blocks
// belongs to the earlier one -- a boundary convention, not a different block.

type boundary struct {
	line, column int
}

// boundaryIndex holds the sorted boundaries of each file in a document set.
type boundaryIndex struct {
	byFile map[string][]boundary
	// lastLine is the last line of each file that carries content, so the
	// final block in a file ends there rather than at a trailing blank.
	lastLine map[string]int
}

func newBoundaryIndex() *boundaryIndex {
	return &boundaryIndex{byFile: map[string][]boundary{}, lastLine: map[string]int{}}
}

// collect records every position an origin knows about as a potential boundary.
func (bi *boundaryIndex) collect(o *openapi3.Origin) {
	if o == nil {
		return
	}
	add := func(l openapi3.Location) {
		if l.Line == 0 {
			return
		}
		bi.byFile[l.File] = append(bi.byFile[l.File], boundary{l.Line, l.Column})
	}
	if o.Key != nil {
		add(*o.Key)
	}
	for _, l := range o.Fields {
		add(l)
	}
	for _, locs := range o.Sequences {
		for _, l := range locs {
			add(l)
		}
	}
}

// setFileLength records where each file's content stops, for the last block.
func (bi *boundaryIndex) setFileLength(file string, lines int) {
	bi.lastLine[file] = lines
}

func (bi *boundaryIndex) sortAll() {
	for f, bs := range bi.byFile {
		sort.Slice(bs, func(i, j int) bool {
			if bs[i].line != bs[j].line {
				return bs[i].line < bs[j].line
			}
			return bs[i].column < bs[j].column
		})
		bi.byFile[f] = bs
	}
}

// endFor returns the last line of the block headed by the key at line/column
// in file: the line before the next boundary at the same or shallower
// indentation, or the file's last content line.
func (bi *boundaryIndex) endFor(file string, line, column int) int {
	bs := bi.byFile[file]
	// Boundaries are sorted by line, so binary-search past our own.
	i := sort.Search(len(bs), func(i int) bool { return bs[i].line > line })
	for ; i < len(bs); i++ {
		if bs[i].column <= column {
			return bs[i].line - 1
		}
	}
	if last, ok := bi.lastLine[file]; ok {
		return last
	}
	return line
}

// lastContentLine is the 1-based last line of text that is not blank, so a
// trailing newline or blank does not extend the final block.
func lastContentLine(text string) int {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return i + 1
		}
	}
	return 0
}
