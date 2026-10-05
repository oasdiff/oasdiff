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
	references := map[*diff.SchemaDiff]schemawalk.References{}
	result := make(Changes, 0, len(changes))
	for _, change := range changes {
		apiChange, ok := change.(ApiChange)
		if !ok {
			result = append(result, change)
			continue
		}
		root, path := apiChange.root, apiChange.propertyPath
		if root == nil || path == "" {
			result = append(result, apiChange)
			continue
		}

		rootReferences, ok := references[root]
		if !ok {
			rootReferences = schemawalk.NewReferences(root)
			references[root] = rootReferences
		}
		name, paths, ok := rootReferences.At(path)
		if !ok {
			result = append(result, apiChange)
			continue
		}

		// Each copy replaces the property path in the message arguments with
		// its own path. If no argument is the property path, the copies would
		// be identical, so the change is reported once with the list of
		// properties.
		argument := slices.IndexFunc(apiChange.Args, func(arg any) bool { return interfaceToString(arg) == path })
		if argument < 0 {
			apiChange.sharedSchema = &SharedSchema{Name: name, Properties: paths}
			result = append(result, apiChange)
			continue
		}
		for _, at := range paths {
			copied := apiChange
			copied.Args = slices.Clone(apiChange.Args)
			copied.Args[argument] = at
			copied.sharedSchema = &SharedSchema{Name: name, Properties: append([]string{at}, slices.DeleteFunc(slices.Clone(paths), func(p string) bool { return p == at })...)}
			result = append(result, copied)
		}
	}
	return result
}
