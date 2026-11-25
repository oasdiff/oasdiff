package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

// subschemaKeyword is a schema keyword whose value is a schema, such as `not`.
type subschemaKeyword struct {
	keyword string // as written in a schema, e.g. "propertyNames"
	idName  string // id segment, e.g. "property-names"
	diff    func(*diff.SchemaDiff) *diff.SchemaDiff
}

var (
	notKeyword = subschemaKeyword{"not", "not", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.NotDiff }}

	propertyNamesKeyword = subschemaKeyword{"propertyNames", "property-names", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.PropertyNamesDiff }}

	unevaluatedItemsKeyword      = subschemaKeyword{"unevaluatedItems", "unevaluated-items", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.UnevaluatedItemsDiff }}
	unevaluatedPropertiesKeyword = subschemaKeyword{"unevaluatedProperties", "unevaluated-properties", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.UnevaluatedPropertiesDiff }}

	ifKeyword   = subschemaKeyword{"if", "if", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.IfDiff }}
	thenKeyword = subschemaKeyword{"then", "then", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ThenDiff }}
	elseKeyword = subschemaKeyword{"else", "else", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ElseDiff }}

	containsKeyword      = subschemaKeyword{"contains", "contains", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ContainsDiff }}
	contentSchemaKeyword = subschemaKeyword{"contentSchema", "content-schema", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ContentSchemaDiff }}
)

// subschemaKeywordChanges reports each of keywords that was added to or
// removed from the body schema of one media type and from its modified
// properties. The ids are <direction>-<body|property>-<idName>-<added|removed>.
func subschemaKeywordChanges(info mediaTypeInfo, direction Direction, keywords ...subschemaKeyword) Changes {
	result := bodySubschemaKeywordChanges(info, direction, keywords...)

	info.walkProperties(func(p propertyInfo) {
		result = append(result, propertySubschemaKeywordChanges(p, direction, keywords...)...)
	})

	return result
}

func bodySubschemaKeywordChanges(info mediaTypeInfo, direction Direction, keywords ...subschemaKeyword) Changes {
	var args []any
	if direction == DirectionResponse {
		args = []any{info.responseStatus}
	}
	return schemaSubschemaKeywordChanges(info, info.schemaDiff, direction, "body", args, info.newChange, keywords)
}

func propertySubschemaKeywordChanges(p propertyInfo, direction Direction, keywords ...subschemaKeyword) Changes {
	args := []any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName)}
	if direction == DirectionResponse {
		args = append(args, p.responseStatus)
	}
	return schemaSubschemaKeywordChanges(p.mediaTypeInfo, p.propertyDiff, direction, "property", args, p.newChange, keywords)
}

func schemaSubschemaKeywordChanges(
	info mediaTypeInfo,
	schemaDiff *diff.SchemaDiff,
	direction Direction,
	scope string,
	args []any,
	newChange func(id string, args []any, comment string) ApiChange,
	keywords []subschemaKeyword,
) Changes {
	result := make(Changes, 0)
	for _, keyword := range keywords {
		keywordDiff := keyword.diff(schemaDiff)
		if keywordDiff == nil {
			continue
		}
		baseSource, revisionSource := location.SchemaFieldSources(info.operationsSources, info.operationItem, schemaDiff, keyword.keyword)
		if keywordDiff.SchemaAdded {
			result = append(result, newChange(boundRuleId(direction, scope, keyword.idName, "added"), args, "").
				WithSources(nil, revisionSource))
		}
		if keywordDiff.SchemaDeleted {
			result = append(result, newChange(boundRuleId(direction, scope, keyword.idName, "removed"), args, "").
				WithSources(baseSource, nil))
		}
	}
	return result
}
