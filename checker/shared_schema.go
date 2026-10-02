package checker

import "github.com/oasdiff/oasdiff/diff"

// SharedSchemaCommentId explains a change reported at one property path when
// several of the payload's paths reach the same schema.
const SharedSchemaCommentId = "shared-schema-comment"

// WithSharedSchema names the schema several of the payload's property paths
// reach, when the change is in it or below it. Such a change is reported at
// one of those paths, so the path no longer tells the reader which schema
// changed; this schema's name does, and the comment says why the other paths
// are absent. A nil shared schema leaves the change alone, and a schema that
// is not a components.schemas entry has no name to give, leaving the comment
// on its own.
func (c ApiChange) WithSharedSchema(shared *diff.SchemaDiff) ApiChange {
	if shared == nil {
		return c
	}
	if c.Comment == "" {
		c.Comment = SharedSchemaCommentId
	}
	c.sharedSchemaName = schemaNameDetail(shared)
	return c
}

// schemaNameDetail formats a schema's components.schemas name as a message
// detail. The revision's name is preferred: it is the one a reader will find
// in the spec they are reviewing.
func schemaNameDetail(schemaDiff *diff.SchemaDiff) string {
	if schemaDiff == nil {
		return ""
	}

	name := schemaDiff.RevisionComponent
	if name == "" {
		name = schemaDiff.BaseComponent
	}
	if name == "" {
		return ""
	}

	return "(shared schema: " + name + ")"
}
