package metaschema

import (
	"encoding"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Action is a syntactic edit applicable at a location: what can literally
// happen to the field in the document, independent of whether it breaks
// clients.
type Action string

const (
	ActionAdd      Action = "add"      // add a map entry, list element, or set member
	ActionRemove   Action = "remove"   // remove a map entry, list element, or set member
	ActionSet      Action = "set"      // the field appears where it was absent
	ActionUnset    Action = "unset"    // the field disappears
	ActionChange   Action = "change"   // the field's value changes to another value
	ActionIncrease Action = "increase" // an ordered value grows
	ActionDecrease Action = "decrease" // an ordered value shrinks
)

// Polarity is the syntactic position of a location in the document.
type Polarity string

const (
	PolarityRequest  Polarity = "request"
	PolarityResponse Polarity = "response"
	PolarityShared   Polarity = "shared"   // components: request or response depending on the referencing site
	PolarityDocument Polarity = "document" // neither wire direction
)

// Cell is one coordinate of the edit space: Action applied at Location.
type Cell struct {
	Location   string
	Action     Action
	Polarity   Polarity
	Annotation bool // spec-defined metadata (description, summary, ...) with no effect on accepted payloads
	Extension  bool // an x-* specification extension
}

// Cube enumerates every cell of the edit space, sorted by location then
// action.
func Cube() []Cell {
	w := &walker{cells: map[Cell]struct{}{}}
	w.walkType(reflect.TypeFor[openapi3.T](), "", false)

	cells := make([]Cell, 0, len(w.cells))
	for c := range w.cells {
		cells = append(cells, c)
	}
	sort.Slice(cells, func(i, j int) bool {
		if cells[i].Location != cells[j].Location {
			return cells[i].Location < cells[j].Location
		}
		return cells[i].Action < cells[j].Action
	})
	return cells
}

// annotationFields are spec-defined metadata: editing them never changes
// which payloads are valid.
var annotationFields = map[string]bool{
	"description":    true,
	"summary":        true,
	"title":          true,
	"example":        true,
	"examples":       true,
	"externalDocs":   true,
	"xml":            true,
	"termsOfService": true,
	"contact":        true,
	"license":        true,
	"tags":           true,
}

var (
	textMarshalerType  = reflect.TypeFor[encoding.TextMarshaler]()
	operationType      = reflect.TypeFor[openapi3.Operation]()
	pathItemType       = reflect.TypeFor[openapi3.PathItem]()
	boolSchemaType     = reflect.TypeFor[openapi3.BoolSchema]()
	exclusiveBoundType = reflect.TypeFor[openapi3.ExclusiveBound]()
	typesType          = reflect.TypeFor[openapi3.Types]()
)

type walker struct {
	cells map[Cell]struct{}
	stack []reflect.Type
}

func (w *walker) emit(path string, annotation, extension bool, actions ...Action) {
	for _, a := range actions {
		w.cells[Cell{
			Location:   path,
			Action:     a,
			Polarity:   polarity(path),
			Annotation: annotation,
			Extension:  extension,
		}] = struct{}{}
	}
}

func polarity(path string) Polarity {
	p := PolarityDocument
	if strings.HasPrefix(path, "components.") {
		p = PolarityShared
	}
	for seg := range strings.SplitSeq(path, ".") {
		switch seg {
		case "parameters", "requestBody":
			p = PolarityRequest
		case "responses":
			p = PolarityResponse
		}
	}
	return p
}

func join(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

func (w *walker) walkType(t reflect.Type, path string, annotation bool) {
	wasPtr := t.Kind() == reflect.Pointer
	if wasPtr {
		t = t.Elem()
	}

	switch {
	case reflect.PointerTo(t).Implements(textMarshalerType):
		// serialized as a plain string (e.g. MappingRef)
		w.emit(path, annotation, false, ActionSet, ActionUnset, ActionChange)
		return
	case isRef(t):
		value, _ := t.FieldByName("Value")
		w.walkType(value.Type, path, annotation)
		return
	case t == boolSchemaType:
		w.emit(path, annotation, false, ActionSet, ActionUnset, ActionChange)
		w.walkType(reflect.TypeFor[openapi3.Schema](), path, annotation)
		return
	case t == exclusiveBoundType:
		w.emit(path, annotation, false, ActionSet, ActionUnset, ActionIncrease, ActionDecrease)
		return
	case t == typesType:
		w.emit(path, annotation, false, ActionAdd, ActionRemove)
		return
	}

	if elem, ok := mapWrapperElem(t); ok {
		entry := join(path, "*")
		w.emit(entry, annotation, false, ActionAdd, ActionRemove)
		w.walkType(elem, entry, annotation)
		return
	}

	switch t.Kind() {
	case reflect.Struct:
		// a map or list entry's existence is already add/remove
		if wasPtr && path != "" && !strings.HasSuffix(path, ".*") {
			w.emit(path, annotation, false, ActionSet, ActionUnset)
		}
		w.walkStruct(t, path, annotation)
	case reflect.Map:
		entry := join(path, "*")
		w.emit(entry, annotation, false, ActionAdd, ActionRemove)
		w.walkType(t.Elem(), entry, annotation)
	case reflect.Slice:
		elem := t.Elem()
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Struct || elem.Kind() == reflect.Map || elem.Kind() == reflect.Slice {
			entry := join(path, "*")
			w.emit(entry, annotation, false, ActionAdd, ActionRemove)
			w.walkType(t.Elem(), entry, annotation)
		} else {
			// a list of scalars is a member set (required, enum, scopes)
			w.emit(path, annotation, false, ActionAdd, ActionRemove)
		}
	case reflect.String, reflect.Interface:
		w.emit(path, annotation, false, ActionSet, ActionUnset, ActionChange)
	case reflect.Bool:
		w.emit(path, annotation, false, ActionSet, ActionUnset)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		w.emit(path, annotation, false, ActionSet, ActionUnset, ActionIncrease, ActionDecrease)
	}
}

func (w *walker) walkStruct(t reflect.Type, path string, annotation bool) {
	if slices.Contains(w.stack, t) {
		return
	}
	w.stack = append(w.stack, t)
	defer func() { w.stack = w.stack[:len(w.stack)-1] }()

	methodDone := false
	for f := range t.Fields() {
		if !f.IsExported() {
			continue
		}
		if f.Name == "Origin" {
			continue
		}
		if f.Name == "Extensions" {
			w.emit(join(path, "x-*"), false, true, ActionAdd, ActionRemove, ActionChange)
			continue
		}
		name := jsonName(f)
		if name == "" {
			// an untagged embedded struct's fields are inlined (e.g. Header
			// embeds Parameter)
			if f.Anonymous {
				w.walkType(f.Type, path, annotation)
			}
			continue
		}
		// the HTTP-method fields of a path item, and the additionalOperations
		// map, all hold an operation keyed by method; they collapse into one
		// wildcard segment: paths.*.* is the operation
		if t == pathItemType && (derefType(f.Type) == operationType ||
			(f.Type.Kind() == reflect.Map && derefType(f.Type.Elem()) == operationType)) {
			if !methodDone {
				methodDone = true
				entry := join(path, "*")
				w.emit(entry, annotation, false, ActionAdd, ActionRemove)
				w.walkType(reflect.PointerTo(operationType), entry, annotation)
			}
			continue
		}
		w.walkType(f.Type, join(path, name), annotation || annotationFields[name])
	}
}

// jsonName returns the field's name in the document, taken from its json
// (or, failing that, yaml) tag; "" means the field is not part of the
// document.
func jsonName(f reflect.StructField) string {
	for _, tag := range []string{"json", "yaml"} {
		v, ok := f.Tag.Lookup(tag)
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(v, ",")
		if name == "-" {
			return ""
		}
		if name != "" {
			return name
		}
	}
	return ""
}

func derefType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}
	return t
}

// isRef reports whether t is a kin-openapi reference wrapper
// ({Ref string, Value *X}); the walk treats it as its resolved value.
func isRef(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	ref, ok := t.FieldByName("Ref")
	if !ok || ref.Type.Kind() != reflect.String {
		return false
	}
	value, ok := t.FieldByName("Value")
	return ok && value.Type.Kind() == reflect.Pointer
}

// mapWrapperElem detects kin-openapi types that hold an unexported map
// behind a Map() accessor (Paths, Responses, Callback) and returns the
// map's element type.
func mapWrapperElem(t reflect.Type) (reflect.Type, bool) {
	m, ok := reflect.PointerTo(t).MethodByName("Map")
	if !ok {
		return nil, false
	}
	// receiver only, one map result
	if m.Type.NumIn() != 1 || m.Type.NumOut() != 1 || m.Type.Out(0).Kind() != reflect.Map {
		return nil, false
	}
	return m.Type.Out(0).Elem(), true
}

// MatchLocation reports whether a rule's location pattern matches a cube
// location. Pattern segments: "**" matches any run of segments (including
// none), "*" matches exactly one segment, anything else matches literally
// (so "*" in a cube location is matched by "*" or "**" in the pattern).
func MatchLocation(pattern, location string) bool {
	return matchSegments(strings.Split(pattern, "."), strings.Split(location, "."))
}

func matchSegments(pat, loc []string) bool {
	if len(pat) == 0 {
		return len(loc) == 0
	}
	if pat[0] == "**" {
		if matchSegments(pat[1:], loc) {
			return true
		}
		if len(loc) == 0 {
			return false
		}
		return matchSegments(pat, loc[1:])
	}
	if len(loc) == 0 {
		return false
	}
	if pat[0] != "*" && pat[0] != loc[0] {
		return false
	}
	return matchSegments(pat[1:], loc[1:])
}
