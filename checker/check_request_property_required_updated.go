package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyBecameRequiredId            = "request-property-became-required"
	RequestPropertyBecameRequiredWithDefaultId = "request-property-became-required-with-default"
	RequestPropertyBecameOptionalId            = "request-property-became-optional"
)

func RequestPropertyRequiredUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		processRequiredDiff := func(schemaDiff *diff.SchemaDiff, propertyPath string, propertyName string) {
			if schemaDiff.RequiredDiff == nil {
				return
			}
			newChange := func(id string, args []any, comment string) ApiChange {
				return info.newChange(id, args, comment).WithSchema(info.schemaDiff, schemaDiff, schemawalk.PropertyFullName(propertyPath, propertyName))
			}
			for _, changedRequiredPropertyName := range schemaDiff.RequiredDiff.Added {
				property, relevant := changedRequiredProperty(schemaDiff, changedRequiredPropertyName, true)
				if !relevant {
					continue
				}
				srcBase, srcRevision := location.SchemaAddedItemSources(operationsSources, info.operationItem, schemaDiff, "required", changedRequiredPropertyName)
				args := []any{schemawalk.PropertyFullName(propertyPath, schemawalk.PropertyFullName(propertyName, changedRequiredPropertyName))}
				if !property.hasDefault {
					result = append(result, newChange(RequestPropertyBecameRequiredId, args, "").
						WithSources(srcBase, srcRevision))
				} else {
					// The property has a default value, but a request that omits a
					// required property is still invalid under the new contract (the
					// default is a server-side fallback, not a validity rule). So this
					// is breaking, same as the no-default case, with a comment
					// explaining why the default does not make it safe.
					result = append(result, newChange(RequestPropertyBecameRequiredWithDefaultId, args, RequiredRequestPropertyWithDefaultCommentId).
						WithSources(srcBase, srcRevision))
				}
			}
			for _, changedRequiredPropertyName := range schemaDiff.RequiredDiff.Deleted {
				if _, relevant := changedRequiredProperty(schemaDiff, changedRequiredPropertyName, false); !relevant {
					continue
				}
				srcBase, srcRevision := location.SchemaDeletedItemSources(operationsSources, info.operationItem, schemaDiff, "required", changedRequiredPropertyName)
				args := []any{schemawalk.PropertyFullName(propertyPath, schemawalk.PropertyFullName(propertyName, changedRequiredPropertyName))}
				result = append(result, newChange(RequestPropertyBecameOptionalId, args, "").
					WithSources(srcBase, srcRevision))
			}
		}

		processRequiredDiff(info.schemaDiff, "", "")

		info.walkProperties(func(p propertyInfo) {
			processRequiredDiff(p.propertyDiff, p.propertyPath, p.propertyName)
		})
	})

	return result
}

func changedRequiredProperty(schemaDiff *diff.SchemaDiff, name string, added bool) (requestProperty, bool) {
	property, relevant := effectiveRequiredChange(schemaDiff, name, added)
	if !relevant || schemaDiff.AllOfDiff == nil {
		return property, relevant
	}
	// Prefer the branch's more specific location when it will report the
	// same edit. A sibling branch may own the declaration instead, in which
	// case only the outer schema has enough information to report it.
	return property, !branchReportsRequiredChange(schemaDiff, name, added, map[*diff.SchemaDiff]bool{})
}

func branchReportsRequiredChange(schemaDiff *diff.SchemaDiff, name string, added bool, seen map[*diff.SchemaDiff]bool) bool {
	if schemaDiff.AllOfDiff == nil || seen[schemaDiff] {
		return false
	}
	seen[schemaDiff] = true
	for _, branch := range schemaDiff.AllOfDiff.Modified {
		d := branch.Diff
		if d.RequiredDiff != nil {
			names := d.RequiredDiff.Added
			if !added {
				names = d.RequiredDiff.Deleted
			}
			if slices.Contains(names, name) {
				if _, relevant := effectiveRequiredChange(d, name, added); relevant {
					return true
				}
			}
		}
		if branchReportsRequiredChange(d, name, added, seen) {
			return true
		}
	}
	return false
}

func effectiveRequiredChange(schemaDiff *diff.SchemaDiff, name string, added bool) (requestProperty, bool) {
	base := lookupRequestProperty(schemaDiff.Base, name)
	revision := lookupRequestProperty(schemaDiff.Revision, name)
	// Property additions and removals have their own checks. A change to an
	// outer required list can also repeat a requirement from an allOf branch.
	if !base.declared || !revision.declared || revision.readOnly {
		return revision, false
	}
	if added {
		return revision, !base.requiredByAllOf
	}
	return revision, !revision.requiredByAllOf
}

type requestProperty struct {
	declared, requiredByAllOf, readOnly, hasDefault bool
}

// Every allOf branch applies to the same object. A property's declaration,
// annotations and required list can therefore live in different branches.
func lookupRequestProperty(schema *openapi3.Schema, name string) (property requestProperty) {
	seen := map[*openapi3.Schema]bool{}
	pending := []*openapi3.Schema{schema}
	for len(pending) > 0 {
		last := len(pending) - 1
		current := pending[last]
		pending = pending[:last]
		if current == nil || seen[current] {
			continue
		}
		seen[current] = true
		if current != schema && slices.Contains(current.Required, name) {
			property.requiredByAllOf = true
		}
		if ref := current.Properties[name]; ref != nil && ref.Value != nil {
			property.declared = true
			property.readOnly = property.readOnly || ref.Value.ReadOnly
			property.hasDefault = property.hasDefault || ref.Value.Default != nil
		}
		for _, branch := range current.AllOf {
			if branch != nil {
				pending = append(pending, branch.Value)
			}
		}
	}
	return property
}
