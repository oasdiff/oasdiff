package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

// singleSchemaKeyword is a schema keyword whose value is one schema, such as
// `not`. Keywords that hold a list or a map of schemas, such as `allOf` and
// `patternProperties`, are not of this kind.
type singleSchemaKeyword struct {
	keyword string // as written in a schema, e.g. "propertyNames"
	idName  string // id segment, e.g. "property-names"
	diff    func(*diff.SchemaDiff) *diff.SchemaDiff
}

var (
	notKeyword = singleSchemaKeyword{"not", "not", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.NotDiff }}

	propertyNamesKeyword = singleSchemaKeyword{"propertyNames", "property-names", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.PropertyNamesDiff }}

	unevaluatedItemsKeyword      = singleSchemaKeyword{"unevaluatedItems", "unevaluated-items", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.UnevaluatedItemsDiff }}
	unevaluatedPropertiesKeyword = singleSchemaKeyword{"unevaluatedProperties", "unevaluated-properties", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.UnevaluatedPropertiesDiff }}

	ifKeyword   = singleSchemaKeyword{"if", "if", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.IfDiff }}
	thenKeyword = singleSchemaKeyword{"then", "then", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ThenDiff }}
	elseKeyword = singleSchemaKeyword{"else", "else", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ElseDiff }}

	containsKeyword      = singleSchemaKeyword{"contains", "contains", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ContainsDiff }}
	contentSchemaKeyword = singleSchemaKeyword{"contentSchema", "content-schema", func(d *diff.SchemaDiff) *diff.SchemaDiff { return d.ContentSchemaDiff }}
)

// singleSchemaKeywordChanges reports each of keywords that was added to or
// removed from the body schema of one media type and from its modified
// properties. The ids are <direction>-<body|property>-<idName>-<added|removed>.
func singleSchemaKeywordChanges(info mediaTypeInfo, direction Direction, keywords ...singleSchemaKeyword) Changes {
	result := bodySingleSchemaKeywordChanges(info, direction, keywords...)

	info.walkProperties(func(p propertyInfo) {
		result = append(result, propertySingleSchemaKeywordChanges(p, direction, keywords...)...)
	})

	return result
}

func bodySingleSchemaKeywordChanges(info mediaTypeInfo, direction Direction, keywords ...singleSchemaKeyword) Changes {
	var args []any
	if direction == DirectionResponse {
		args = []any{info.responseStatus}
	}
	return nodeSingleSchemaKeywordChanges(info, info.schemaDiff, direction, "body", args, info.newChange, keywords)
}

func propertySingleSchemaKeywordChanges(p propertyInfo, direction Direction, keywords ...singleSchemaKeyword) Changes {
	args := []any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName)}
	if direction == DirectionResponse {
		args = append(args, p.responseStatus)
	}
	return nodeSingleSchemaKeywordChanges(p.mediaTypeInfo, p.propertyDiff, direction, "property", args, p.newChange, keywords)
}

func nodeSingleSchemaKeywordChanges(
	info mediaTypeInfo,
	schemaDiff *diff.SchemaDiff,
	direction Direction,
	scope string,
	args []any,
	newChange func(id string, args []any, comment string) ApiChange,
	keywords []singleSchemaKeyword,
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
