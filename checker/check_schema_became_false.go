package checker

import (
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

		info.walkProperties(func(p propertyInfo) {
			if id := falseSchemaChangeId(p.propertyDiff, RequestPropertySchemaBecameFalseId, RequestPropertySchemaBecameNotFalseId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName)},
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

		info.walkProperties(func(p propertyInfo) {
			if id := falseSchemaChangeId(p.propertyDiff, ResponsePropertySchemaBecameFalseId, ResponsePropertySchemaBecameNotFalseId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "type")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					falseSchemaComment(id),
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}

func RequestParameterSchemaBecameFalseCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ParametersDiff == nil || operationItem.ParametersDiff.Modified == nil {
				continue
			}
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for paramLocation, paramItems := range operationItem.ParametersDiff.Modified {
				for paramName, paramItem := range paramItems {
					if paramItem.SchemaDiff == nil {
						continue
					}

					if id := falseSchemaChangeId(paramItem.SchemaDiff, RequestParameterSchemaBecameFalseId, RequestParameterSchemaBecameNotFalseId); id != "" {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramItem.SchemaDiff, "type")
						result = append(result, opInfo.NewApiChange(
							id,
							[]any{paramLocation, paramName},
							"",
						).WithSchema(paramItem.SchemaDiff).WithSources(baseSource, revisionSource))
					}

					checkModifiedPropertiesDiff(
						paramItem.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if id := falseSchemaChangeId(propertyDiff, RequestParameterPropertySchemaBecameFalseId, RequestParameterPropertySchemaBecameNotFalseId); id != "" {
								baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "type")
								result = append(result, opInfo.NewApiChange(
									id,
									[]any{propertyFullName(propertyPath, propertyName), paramLocation, paramName},
									"",
								).WithSchema(propertyDiff).WithSources(baseSource, revisionSource))
							}
						})
				}
			}
		}
	}
	return result
}

func ResponseHeaderSchemaBecameFalseCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.HeadersDiff == nil {
					continue
				}
				for headerName, headerDiff := range responseDiff.HeadersDiff.Modified {
					if headerDiff.SchemaDiff == nil {
						continue
					}
					if id := falseSchemaChangeId(headerDiff.SchemaDiff, ResponseHeaderSchemaBecameFalseId, ResponseHeaderSchemaBecameNotFalseId); id != "" {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, headerDiff.SchemaDiff, "type")
						result = append(result, opInfo.NewApiChange(
							id,
							[]any{headerName, responseStatus},
							"",
						).WithSchema(headerDiff.SchemaDiff).WithSources(baseSource, revisionSource))
					}
				}
			}
		}
	}
	return result
}
