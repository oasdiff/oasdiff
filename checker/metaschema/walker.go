package metaschema

import (
	"encoding"
	"reflect"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

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
	edits map[Edit]struct{}
	stack []reflect.Type
}

func (w *walker) emit(path string, annotation, extension bool, actions ...Action) {
	for _, a := range actions {
		w.edits[Edit{
			Location:   path,
			Action:     a,
			Polarity:   polarity(path),
			Annotation: annotation,
			Extension:  extension,
		}] = struct{}{}
	}
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
		entry := joinLocation(path, "*")
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
		entry := joinLocation(path, "*")
		w.emit(entry, annotation, false, ActionAdd, ActionRemove)
		w.walkType(t.Elem(), entry, annotation)
	case reflect.Slice:
		elem := t.Elem()
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Struct || elem.Kind() == reflect.Map || elem.Kind() == reflect.Slice {
			entry := joinLocation(path, "*")
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
		// kin holds specification extensions in a map whose json tag hides
		// it, so this is matched by name and shape rather than by the tag
		// that names every other field
		if f.Name == "Extensions" && f.Type.Kind() == reflect.Map {
			w.emit(joinLocation(path, "x-*"), false, true, ActionAdd, ActionRemove, ActionChange)
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
				entry := joinLocation(path, "*")
				w.emit(entry, annotation, false, ActionAdd, ActionRemove)
				w.walkType(reflect.PointerTo(operationType), entry, annotation)
			}
			continue
		}
		w.walkType(f.Type, joinLocation(path, name), annotation || annotationFields[name])
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
