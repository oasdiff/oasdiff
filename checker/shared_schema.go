package checker

import (
	"fmt"
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

// SharedSchemaCommentId explains a change reported at one property path when
// several of the payload's paths reach the same schema.
const SharedSchemaCommentId = "shared-schema-comment"

// SharedSchema is the schema several properties of a payload reach, attached
// to a change in it or below it, which is reported at only one of them.
type SharedSchema struct {
	// Name is empty for a schema that is not a components.schemas entry.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Properties starts with the path the change is reported at. The rest
	// reach the same change through the schema's other references, one path
	// per reference.
	Properties []string `json:"properties" yaml:"properties"`
}

// WithSharedSchema attaches the schema several of the payload's properties
// reach. The change is reported at one of them, so it names the schema and
// the other properties, and the comment says why they are not reported
// separately. A nil shared schema leaves the change alone.
func (c ApiChange) WithSharedSchema(shared *SharedSchema) ApiChange {
	if shared == nil {
		return c
	}
	if c.Comment == "" {
		c.Comment = SharedSchemaCommentId
	}
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

// sharedReach is a schema several of the payload's property paths reach, as
// one walk found it.
type sharedReach struct {
	name string
	// path is the one this walk took to the schema.
	path string
	// paths reach the schema, one per reference, in walk order.
	paths []string
}

// at is the shared schema for a change at path, which the walk reached at or
// below the schema: each other reference reaches the same change once the
// part of path below the schema is appended to it.
func (r *sharedReach) at(path string) *SharedSchema {
	if r == nil {
		return nil
	}

	properties := []string{path}
	if below, ok := strings.CutPrefix(path, r.path); ok {
		for _, p := range r.paths {
			if other := p + below; !slices.Contains(properties, other) {
				properties = append(properties, other)
			}
		}
	}

	return &SharedSchema{Name: r.name, Properties: properties}
}

// componentName prefers the revision's name: it is the one a reader will
// find in the spec they are reviewing.
func componentName(schemaDiff *diff.SchemaDiff) string {
	if schemaDiff.RevisionComponent != "" {
		return schemaDiff.RevisionComponent
	}
	return schemaDiff.BaseComponent
}
