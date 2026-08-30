package checker

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodySchemaBecameFalseId        = "request-body-schema-became-false"
	RequestBodySchemaBecameNotFalseId     = "request-body-schema-became-not-false"
	RequestPropertySchemaBecameFalseId    = "request-property-schema-became-false"
	RequestPropertySchemaBecameNotFalseId = "request-property-schema-became-not-false"

	ResponseBodySchemaBecameFalseId        = "response-body-schema-became-false"
	ResponseBodySchemaBecameNotFalseId     = "response-body-schema-became-not-false"
	ResponsePropertySchemaBecameFalseId    = "response-property-schema-became-false"
	ResponsePropertySchemaBecameNotFalseId = "response-property-schema-became-not-false"

	RequestParameterSchemaBecameFalseId            = "request-parameter-schema-became-false"
	RequestParameterSchemaBecameNotFalseId         = "request-parameter-schema-became-not-false"
	RequestParameterPropertySchemaBecameFalseId    = "request-parameter-property-schema-became-false"
	RequestParameterPropertySchemaBecameNotFalseId = "request-parameter-property-schema-became-not-false"

	ResponseHeaderSchemaBecameFalseId    = "response-header-schema-became-false"
	ResponseHeaderSchemaBecameNotFalseId = "response-header-schema-became-not-false"

	// A property schema narrowing to nothing is reported at the level the
	// law derives for a response narrowing, with this comment naming what
	// the reader likely wants to know instead.
	ResponsePropertySchemaBecameFalseCommentId = "response-property-schema-became-false-comment"
)

// falseSchemaChangeId returns the id for a schema that became or stopped
// being the boolean schema `false`, or "" when neither happened. `true` is
// not a transition worth an id of its own: it accepts the same instances as
// the empty schema, so a change to or from it is fully described by the
// keyword-level findings.
func falseSchemaChangeId(d *diff.SchemaDiff, falseId, notFalseId string) string {
	if d == nil || d.AlwaysDiff == nil {
		return ""
	}
	if d.AlwaysDiff.To == false {
		return falseId
	}
	if d.AlwaysDiff.From == false {
		return notFalseId
	}
	return ""
}

// falseItemsChangeId is falseSchemaChangeId for an items schema that appears
// as or disappears from `false` (closing or reopening a tuple). A one-sided
// diff carries only the added or deleted flag and the property walk skips
// one-sided nodes, so the transition is detected at the parent, whose
// document sides hold the schema. Other one-sided sub-schema slots stay
// unclassified (#1199).
func falseItemsChangeId(d *diff.SchemaDiff, falseId, notFalseId string) string {
	if d == nil || d.ItemsDiff == nil {
		return ""
	}
	if d.ItemsDiff.SchemaAdded && d.Revision != nil && isFalseSchemaRef(d.Revision.Items) {
		return falseId
	}
	if d.ItemsDiff.SchemaDeleted && d.Base != nil && isFalseSchemaRef(d.Base.Items) {
		return notFalseId
	}
	return ""
}

func isFalseSchemaRef(ref *openapi3.SchemaRef) bool {
	return ref != nil && ref.Value != nil && ref.Value.Always != nil && !*ref.Value.Always
}

func falseSchemaComment(id string) string {
	if id == ResponsePropertySchemaBecameFalseId {
		return ResponsePropertySchemaBecameFalseCommentId
	}
	return ""
}

func RequestPropertySchemaBecameFalseCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if id := falseSchemaChangeId(info.schemaDiff, RequestBodySchemaBecameFalseId, RequestBodySchemaBecameNotFalseId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "type")
			result = append(result, info.newChange(id, nil, "").
				WithSources(baseSource, revisionSource))
		}

		if id := falseItemsChangeId(info.schemaDiff, RequestPropertySchemaBecameFalseId, RequestPropertySchemaBecameNotFalseId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "items")
			result = append(result, info.newChange(id, []any{"items"}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if id := falseSchemaChangeId(p.propertyDiff, RequestPropertySchemaBecameFalseId, RequestPropertySchemaBecameNotFalseId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName)},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
			if id := falseItemsChangeId(p.propertyDiff, RequestPropertySchemaBecameFalseId, RequestPropertySchemaBecameNotFalseId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "items")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName, "items")},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}

func ResponsePropertySchemaBecameFalseCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if id := falseSchemaChangeId(info.schemaDiff, ResponseBodySchemaBecameFalseId, ResponseBodySchemaBecameNotFalseId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "type")
			result = append(result, info.newChange(id, nil, falseSchemaComment(id)).
				WithSources(baseSource, revisionSource))
		}

		if id := falseItemsChangeId(info.schemaDiff, ResponsePropertySchemaBecameFalseId, ResponsePropertySchemaBecameNotFalseId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "items")
			result = append(result, info.newChange(id, []any{"items", info.responseStatus}, falseSchemaComment(id)).
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if id := falseSchemaChangeId(p.propertyDiff, ResponsePropertySchemaBecameFalseId, ResponsePropertySchemaBecameNotFalseId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					falseSchemaComment(id),
				).WithSources(propBaseSource, propRevisionSource))
			}
			if id := falseItemsChangeId(p.propertyDiff, ResponsePropertySchemaBecameFalseId, ResponsePropertySchemaBecameNotFalseId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "items")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName, "items"), info.responseStatus},
					falseSchemaComment(id),
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}

func RequestParameterSchemaBecameFalseCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		if id := falseSchemaChangeId(p.paramDiff.SchemaDiff, RequestParameterSchemaBecameFalseId, RequestParameterSchemaBecameNotFalseId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "type")
			result = append(result, p.opInfo.NewApiChange(
				id,
				[]any{p.location, p.name},
				"",
			).WithSchema(p.paramDiff.SchemaDiff).WithSources(baseSource, revisionSource))
		}

		checkModifiedPropertiesDiff(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
				if id := falseSchemaChangeId(propertyDiff, RequestParameterPropertySchemaBecameFalseId, RequestParameterPropertySchemaBecameNotFalseId); id != "" {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, propertyDiff, "type")
					result = append(result, p.opInfo.NewApiChange(
						id,
						[]any{propertyFullName(propertyPath, propertyName), p.location, p.name},
						"",
					).WithSchema(propertyDiff).WithSources(baseSource, revisionSource))
				}
			})
	})

	return result
}

func ResponseHeaderSchemaBecameFalseCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseHeaders(diffReport, operationsSources, config, func(h headerInfo) {
		if h.headerDiff.SchemaDiff == nil {
			return
		}
		if id := falseSchemaChangeId(h.headerDiff.SchemaDiff, ResponseHeaderSchemaBecameFalseId, ResponseHeaderSchemaBecameNotFalseId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, h.opInfo.methodDiff, h.headerDiff.SchemaDiff, "type")
			result = append(result, h.opInfo.NewApiChange(
				id,
				[]any{h.name, h.responseStatus},
				"",
			).WithSchema(h.headerDiff.SchemaDiff).WithSources(baseSource, revisionSource))
		}
	})

	return result
}
