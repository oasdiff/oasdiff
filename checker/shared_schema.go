package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

// SharedSchemaCommentId explains a change reported at one property when
// several properties of the payload reach the schema it is in.
const SharedSchemaCommentId = "shared-schema-comment"

const (
	SharedSchemaDetailNameId       = "shared-schema-detail-name"
	SharedSchemaDetailAlsoId       = "shared-schema-detail-also"
	SharedSchemaDetailAlsoMoreId   = "shared-schema-detail-also-more"
	SharedSchemaDetailAlsoCyclicId = "shared-schema-detail-also-cyclic"
)

// sharedSchemaProperties is how many properties a shared schema lists. Over
// the APIs-guru corpus and the GitHub REST API description, a change in a
// shared schema is at three properties or fewer in 59% of cases.
const sharedSchemaProperties = 3

// SharedSchema is the schema several properties of a payload reach, attached
// to a change in it or below it.
type SharedSchema struct {
	// Name is empty for a schema that is not a components.schemas entry, and
	// is not inside one.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Properties are the first few properties the change is at, the one it is
	// reported at first.
	Properties []string `json:"properties" yaml:"properties"`
	// Count is the number of properties the change is at. It is zero when
	// Cyclic.
	Count int `json:"count,omitempty" yaml:"count,omitempty"`
	// Cyclic is set when a cycle makes the number of properties unbounded.
	Cyclic bool `json:"cyclic,omitempty" yaml:"cyclic,omitempty"`
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

// detail renders the shared schema as a message detail: the other properties
// it lists, and how many more there are.
func (s *SharedSchema) detail(l Localizer, format func([]any) []any) string {
	if s == nil {
		return ""
	}

	parts := []string{}
	if s.Name != "" {
		parts = append(parts, l(SharedSchemaDetailNameId, format([]any{s.Name})...))
	}
	if others := s.Properties[1:]; len(others) > 0 {
		listed := make([]string, len(others))
		for i, other := range format(toAny(others)) {
			listed[i] = interfaceToString(other)
		}
		list := strings.Join(listed, ", ")
		switch more := s.Count - len(s.Properties); {
		case s.Cyclic:
			parts = append(parts, l(SharedSchemaDetailAlsoCyclicId, list))
		case more > 0:
			parts = append(parts, l(SharedSchemaDetailAlsoMoreId, list, more))
		default:
			parts = append(parts, l(SharedSchemaDetailAlsoId, list))
		}
	}
	if len(parts) == 0 {
		return ""
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

func toAny(values []string) []any {
	result := make([]any, len(values))
	for i, v := range values {
		result[i] = v
	}
	return result
}

// attachSharedSchemas handles a schema that several properties of the same
// payload reference. The walk goes into such a schema only through its first
// reference, so a change inside it is found once, at one property. This names
// the schema on the change and lists the other properties it is at.
func attachSharedSchemas(changes Changes) Changes {
	references := referencesByRoot{}
	for i, change := range changes {
		changes[i] = attachSharedSchema(change, references)
	}
	return changes
}

// attachSharedSchema returns the change with its shared schema, or unchanged
// if it is in none.
func attachSharedSchema(change Change, references referencesByRoot) Change {
	apiChange, ok := change.(ApiChange)
	if !ok || apiChange.root == nil || apiChange.propertyPath == "" {
		return change
	}

	shared, ok := references.of(apiChange.root).At(apiChange.propertyPath, sharedSchemaProperties)
	if !ok {
		return change
	}

	if apiChange.Comment == "" {
		apiChange.Comment = SharedSchemaCommentId
	}
	return apiChange.WithSharedSchema(&SharedSchema{Name: shared.Name, Properties: shared.Paths, Count: shared.Count, Cyclic: shared.Cyclic})
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
