package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseHeaderTypeChangedId     = "response-header-type-changed"
	ResponseHeaderTypeGeneralizedId = "response-header-type-generalized"
	ResponseHeaderTypeSpecializedId = "response-header-type-specialized"
	ResponseHeaderTypeCompatibleId  = "response-header-type-compatible"

	// ResponseHeaderTypeCompatibleCommentId explains the surprising "compatible"
	// verdict for a header, where the shared media-type comment (which talks about
	// XML) does not apply: a header value is text on the wire regardless.
	ResponseHeaderTypeCompatibleCommentId = "response-header-type-compatible-comment"
)

// ResponseHeaderTypeChangedCheck reports a type or format change on a response
// header's schema, the response-header mirror of ResponsePropertyTypeChangedCheck
// (response headers previously had existence-level checks only, no schema walk).
//
// A response header value is text on the wire, never a strongly typed media
// type, so it is classified non-strongly-typed (mediaType ""), the same way a
// scalar parameter is: a bare string -> integer is a loosely typed compatible
// change (the value "123" is still valid text a string client reads), while
// dropping a format a client parses -- the motivating case, Retry-After
// string/date-time -> integer -- is breaking on the format axis.
func ResponseHeaderTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedResponseHeaders(diffReport, operationsSources, config, func(h headerInfo) {
		// Only the simple schema serialization (the common case), where
		// the value is text on the wire and non-strongly-typed. A header
		// serialized via `content` (e.g. content: application/json)
		// produces a ContentDiff, not a SchemaDiff, and would be
		// strongly typed; that is a separate slice, not covered here.
		schemaDiff := h.headerDiff.SchemaDiff
		if schemaDiff == nil || schemaDiff.Base == nil || schemaDiff.Revision == nil {
			return
		}
		typeDiff := schemaDiff.TypeDiff
		formatDiff := schemaDiff.FormatDiff
		if typeDiff.Empty() && formatDiff.Empty() {
			return
		}

		// A header value is text on the wire, never a strongly typed
		// media type, so stronglyTyped is false (like a scalar
		// parameter); the compatible verdict carries a header-specific
		// comment rather than the shared media-type (XML) one.
		id, comment := responseTypeChangeId(typeDiff, formatDiff, false, ResponseHeaderTypeCompatibleCommentId, schemaDiff,
			ResponseHeaderTypeSpecializedId, ResponseHeaderTypeCompatibleId, ResponseHeaderTypeGeneralizedId, ResponseHeaderTypeChangedId)

		baseSource, revisionSource := headerSources(operationsSources, h.opInfo.methodDiff, h.responseDiff, h.name)
		result = append(result, h.opInfo.NewApiChange(
			id,
			[]any{h.name, getTypeFormatDimension(schemaDiff), getBaseTypeFormat(schemaDiff), getRevisionTypeFormat(schemaDiff), h.responseStatus},
			comment,
		).WithSchema(schemaDiff).WithSources(baseSource, revisionSource))
	})
	return result
}
