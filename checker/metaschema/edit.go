package metaschema

import (
	"reflect"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// Edit is one way an OpenAPI document can be modified: an Action applied at
// a Location.
type Edit struct {
	Location   string
	Action     Action
	Polarity   Polarity
	Annotation bool // spec-defined metadata (description, summary, ...) with no effect on accepted payloads
	Extension  bool // an x-* specification extension
}

// Edits enumerates every possible edit of an OpenAPI document, sorted by
// location then action.
func Edits() []Edit {
	w := &walker{edits: map[Edit]struct{}{}}
	w.walkType(reflect.TypeFor[openapi3.T](), "", false)

	edits := make([]Edit, 0, len(w.edits))
	for c := range w.edits {
		edits = append(edits, c)
	}
	sort.Slice(edits, func(i, j int) bool {
		if edits[i].Location != edits[j].Location {
			return edits[i].Location < edits[j].Location
		}
		return edits[i].Action < edits[j].Action
	})
	return edits
}
