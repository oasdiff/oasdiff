package checker

import (
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

// SharedSchemaCommentId explains a change reported at one property when
// several properties of the payload reach the schema it is in.
const SharedSchemaCommentId = "shared-schema-comment"

const (
	SharedSchemaDetailNameId     = "shared-schema-detail-name"
	SharedSchemaDetailAlsoId     = "shared-schema-detail-also"
	SharedSchemaDetailAlsoMoreId = "shared-schema-detail-also-more"
)

// SharedSchema is the schema several properties of a payload reach, attached
// to a change in it or below it.
type SharedSchema struct {
	// Name is empty for a schema that is not a components.schemas entry, and
	// is not inside one.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Properties starts with the path the change is reported at. The rest
	// reach the same change through the schema's other references, one path
	// per reference.
	Properties []string `json:"properties" yaml:"properties"`
}

// WithSharedSchema returns a copy of the ApiChange in a schema several
// properties of its payload reach.
func (c ApiChange) WithSharedSchema(shared *SharedSchema) ApiChange {
	c.sharedSchema = shared
	return c
}

// GetSharedSchema is nil unless several of the payload's properties reach the
// schema the change is in or below.
func (c ApiChange) GetSharedSchema() *SharedSchema {
	return c.sharedSchema
}

// detail renders the shared schema as a message detail. It lists one other
// property and counts the rest: the paths are long, and the full list is in
// Properties.
func (s *SharedSchema) detail(l Localizer, format func([]any) []any) string {
	if s == nil {
		return ""
	}

	parts := []string{}
	if s.Name != "" {
		parts = append(parts, l(SharedSchemaDetailNameId, format([]any{s.Name})...))
	}
	if others := s.Properties[1:]; len(others) > 0 {
		other := format([]any{others[0]})[0]
		if more := len(others) - 1; more > 0 {
			parts = append(parts, l(SharedSchemaDetailAlsoMoreId, other, more))
		} else {
			parts = append(parts, l(SharedSchemaDetailAlsoId, other))
		}
	}
	if len(parts) == 0 {
		return ""
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

// expandSharedSchemas handles a schema that several properties of the same
// payload reference. The walk goes into such a schema only through its first
// reference, so a change inside it is found only there. This reports the
// change once at each reference. Each copy names the schema and the other
// properties.
func expandSharedSchemas(changes Changes) Changes {
	references := referencesByRoot{}
	result := make(Changes, 0, len(changes))
	for _, change := range changes {
		result = append(result, expandSharedSchema(change, references)...)
	}
	return result
}

// expandSharedSchema returns the change once per reference to the shared
// schema it is in or below, or the change alone if it is in none.
func expandSharedSchema(change Change, references referencesByRoot) Changes {
	apiChange, ok := change.(ApiChange)
	if !ok || apiChange.root == nil || apiChange.propertyPath == "" {
		return Changes{change}
	}

	name, paths, ok := references.of(apiChange.root).At(apiChange.propertyPath)
	if !ok {
		return Changes{change}
	}

	argument := propertyArgument(apiChange)
	if argument < 0 {
		// Copies would be identical, so the change is reported once with the
		// list of properties.
		return Changes{apiChange.WithSharedSchema(&SharedSchema{Name: name, Properties: paths})}
	}

	result := make(Changes, 0, len(paths))
	for _, at := range paths {
		result = append(result, copyAt(apiChange, argument, name, paths, at))
	}
	return result
}

// referencesByRoot builds the references of each root once, however many
// changes below it ask.
type referencesByRoot map[*diff.SchemaDiff]schemawalk.References

func (r referencesByRoot) of(root *diff.SchemaDiff) schemawalk.References {
	references, ok := r[root]
	if !ok {
		references = schemawalk.NewReferences(root)
		r[root] = references
	}
	return references
}

// propertyArgument is the index of the message argument that names the
// change's property path, or -1 if no argument does.
func propertyArgument(change ApiChange) int {
	return slices.IndexFunc(change.Args, func(arg any) bool { return interfaceToString(arg) == change.propertyPath })
}

// copyAt returns the change reported at the property path at: the argument
// that named the original path names at, and the shared schema lists at first.
func copyAt(change ApiChange, argument int, name string, paths []string, at string) ApiChange {
	change.Args = slices.Clone(change.Args)
	change.Args[argument] = at
	others := slices.DeleteFunc(slices.Clone(paths), func(p string) bool { return p == at })
	return change.WithSharedSchema(&SharedSchema{Name: name, Properties: append([]string{at}, others...)})
}
