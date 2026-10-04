package checker

import (
	"fmt"
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

// SharedSchemaCommentId explains a change reported at one property when
// several properties of the payload reach the schema it is in.
const SharedSchemaCommentId = "shared-schema-comment"

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
func (s *SharedSchema) detail() string {
	if s == nil {
		return ""
	}

	parts := []string{}
	if s.Name != "" {
		parts = append(parts, "shared schema: "+s.Name)
	}
	if others := s.Properties[1:]; len(others) > 0 {
		also := "also at `" + others[0] + "`"
		if more := len(others) - 1; more > 0 {
			also += fmt.Sprintf(" and %d more", more)
		}
		parts = append(parts, also)
	}
	if len(parts) == 0 {
		return ""
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

// expandSharedSchemas reports a change in or below a schema several
// references of its payload reach at each of those references. The walk
// continued through the first reference only, so the change was found there,
// and without this the others would be missing. Each copy names the schema
// and the other properties.
//
// It is the last step to read where a change was computed, so the changes it
// returns no longer hold on to the diff.
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
		apiChange.root, apiChange.schema, apiChange.propertyPath = nil, nil, ""
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

		// The copy for another reference names its property in the argument
		// that named this one. A message that names no property cannot be
		// told apart from its copies, so it is reported once, listing them.
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
