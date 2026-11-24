package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyAddedToClosedObjectId = "response-property-added-to-closed-object"
)

// ResponsePropertyAddedToClosedObjectCheck reports a property added to a
// response object that was closed in the base spec. Clients that validate
// against the previous schema reject the new key, so unlike an addition to an
// open object this is breaking.
//
// An object is closed by additionalProperties: false or, in OpenAPI 3.1, by
// unevaluatedProperties: false. unevaluatedProperties also sees properties
// evaluated by in-place applicators, so a schema reached through allOf, anyOf
// or oneOf of a schema with unevaluatedProperties: false is closed too, unless
// the base already evaluated the added name elsewhere under that schema.
func ResponsePropertyAddedToClosedObjectCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		enclosing := unevaluatedClosures(info.schemaDiff)
		schemawalk.AddedProperties(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff, underAllOf bool) {
				closedBy := closingKeyword(parent, propertyName, enclosing)
				if closedBy == "" {
					return
				}
				if propertyItem.WriteOnly {
					return
				}
				propName := schemawalk.PropertyFullName(propertyPath, propertyName)
				revisionSource := location.PropertySource(operationsSources, info.operationItem.Revision, propertyItem)
				result = append(result, info.newChange(
					ResponsePropertyAddedToClosedObjectId,
					[]any{propName, info.responseStatus, closedBy},
					"",
				).WithSchema(info.schemaDiff, parent, propName).WithAllOfDisclaimers(underAllOf, nil).WithSources(nil, revisionSource))
			})
	})

	return result
}

const (
	closedByAdditionalProperties  = "additionalProperties: false"
	closedByUnevaluatedProperties = "unevaluatedProperties: false"
)

// closingKeyword returns the keyword that made the base schema reject a key
// named propertyName, newly declared on parent, or "" if the base admitted it.
func closingKeyword(parent *diff.SchemaDiff, propertyName string, enclosing map[*diff.SchemaDiff][]*openapi3.Schema) string {
	if parent == nil {
		return ""
	}
	if getAdditionalPropertiesMode(parent.Base) == additionalPropertiesNone {
		return closedByAdditionalProperties
	}
	if unevaluatedPropertiesClosed(parent.Base) {
		return closedByUnevaluatedProperties
	}
	for _, closure := range enclosing[parent] {
		if !evaluatesProperty(closure, propertyName, map[*openapi3.Schema]struct{}{}) {
			return closedByUnevaluatedProperties
		}
	}
	return ""
}

func unevaluatedPropertiesClosed(schema *openapi3.Schema) bool {
	return schema != nil && schema.UnevaluatedProperties.Has != nil && !*schema.UnevaluatedProperties.Has
}

// unevaluatedClosures maps each schema diff reached from root only through
// allOf, anyOf or oneOf of a schema closed by unevaluatedProperties: false to
// the base of those closing schemas.
func unevaluatedClosures(root *diff.SchemaDiff) map[*diff.SchemaDiff][]*openapi3.Schema {
	result := map[*diff.SchemaDiff][]*openapi3.Schema{}
	if root == nil {
		return result
	}
	mark := func(sd *diff.SchemaDiff) {
		if sd == nil || !unevaluatedPropertiesClosed(sd.Base) {
			return
		}
		closure := sd.Base
		seen := map[*diff.SchemaDiff]struct{}{}
		var visit func(*diff.SchemaDiff)
		visit = func(node *diff.SchemaDiff) {
			for _, subschemas := range []*diff.SubschemasDiff{node.AllOfDiff, node.AnyOfDiff, node.OneOfDiff} {
				if subschemas == nil {
					continue
				}
				for _, modified := range subschemas.Modified {
					child := modified.Diff
					if child == nil {
						continue
					}
					if _, ok := seen[child]; ok {
						continue
					}
					seen[child] = struct{}{}
					result[child] = append(result[child], closure)
					visit(child)
				}
			}
		}
		visit(sd)
	}
	mark(root)
	schemawalk.Walker{Enter: func(_ string, _ string, sd *diff.SchemaDiff, _ *diff.SchemaDiff, _ bool) {
		mark(sd)
	}}.Walk(root)
	return result
}

// evaluatesProperty reports whether schema, or a subschema it applies in
// place through allOf, anyOf or oneOf, declares the property name, so that
// unevaluatedProperties on schema does not apply to it. Branches of anyOf and
// oneOf count even though a given instance may not match them; that can only
// hide a report, never add one.
func evaluatesProperty(schema *openapi3.Schema, name string, visiting map[*openapi3.Schema]struct{}) bool {
	if schema == nil {
		return false
	}
	if _, ok := visiting[schema]; ok {
		return false
	}
	visiting[schema] = struct{}{}
	if _, ok := schema.Properties[name]; ok {
		return true
	}
	for _, refs := range []openapi3.SchemaRefs{schema.AllOf, schema.AnyOf, schema.OneOf} {
		for _, ref := range refs {
			if ref != nil && evaluatesProperty(ref.Value, name, visiting) {
				return true
			}
		}
	}
	return false
}
