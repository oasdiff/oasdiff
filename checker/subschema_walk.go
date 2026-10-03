package checker

import (
	"fmt"
	"maps"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// subschemaWalk recurses over a schema diff's sub-schemas once, so that a walk
// using it supplies only the part that differs: what it emits, and where.
type subschemaWalk struct {
	// enter is called for the node itself, before its name is appended to the
	// path, so a caller receives the two separately.
	enter func(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, underAllOf bool, shared *sharedReach)
	// properties is called for a node's own properties, after its name is
	// appended, so the path already names the node they belong to.
	properties func(propertyPath string, schemaDiff *diff.SchemaDiff, underAllOf bool, shared *sharedReach)
	// A schema reached through several property paths is walked once: a
	// change in it is one change to the operation's contract, however many
	// paths lead to it.
	seen map[walkVisit]struct{}
	// reaches, when set, records the path of every reference to each schema
	// instead of reporting anything (sharedSchemas).
	reaches map[walkVisit][]string
	// shared holds the paths to each schema more than one reference reaches,
	// from a pass over the same diff (sharedSchemas).
	shared map[walkVisit][]string
	// sharedAncestor is the innermost schema at or above the node being
	// walked that several references reach. The hooks receive it: a change
	// there or below it is reported at one path out of several.
	sharedAncestor *sharedReach
	// namedAncestor is the innermost components.schemas entry at or above the
	// node being walked. It names a shared schema written inline, which is
	// shared because the entry it belongs to was copied, as the parser does
	// for a $ref with a description beside it, and the copies keep its
	// children.
	namedAncestor string
}

type walkVisit struct {
	schemaDiff *diff.SchemaDiff
	underAllOf bool
}

// sharedSchemas returns the paths to each schema below schemaDiff that more
// than one reference reaches, which is what the walk itself then reports only
// once. A reference is recorded at the path the walk takes to it, so the
// count is of references rather than of every chain that leads to one.
func sharedSchemas(schemaDiff *diff.SchemaDiff, underAllOf bool) map[walkVisit][]string {
	reaches := map[walkVisit][]string{}
	subschemaWalk{reaches: reaches}.walk("", "", schemaDiff, nil, underAllOf)

	shared := map[walkVisit][]string{}
	for visit, paths := range reaches {
		if len(paths) > 1 {
			shared[visit] = paths
		}
	}
	return shared
}

func (w subschemaWalk) walk(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, underAllOf bool) {
	if w.seen == nil {
		w.seen = make(map[walkVisit]struct{})
	}
	nodePath := propertyPath
	if propertyName != "" {
		nodePath = propertyFullName(propertyPath, propertyName)
	}

	visit := walkVisit{schemaDiff: schemaDiff, underAllOf: underAllOf}
	if w.reaches != nil {
		w.reaches[visit] = append(w.reaches[visit], nodePath)
	}
	if _, ok := w.seen[visit]; ok {
		return
	}
	w.seen[visit] = struct{}{}

	// Assigning to the value receiver scopes these to the subtree: every
	// recursive call below inherits them, siblings do not.
	if name := componentName(schemaDiff); name != "" {
		w.namedAncestor = name
	}
	if paths, ok := w.shared[visit]; ok {
		w.sharedAncestor = &sharedReach{name: w.namedAncestor, path: nodePath, paths: paths}
	}

	if w.enter != nil && (propertyName != "" || propertyPath != "") {
		w.enter(propertyPath, propertyName, schemaDiff, parentDiff, underAllOf, w.sharedAncestor)
	}

	propertyPath = nodePath

	if schemaDiff.AllOfDiff != nil {
		for _, v := range schemaDiff.AllOfDiff.Modified {
			w.walk(propertyFullName(propertyPath, fmt.Sprintf("allOf[%s]", v)), "", v.Diff, schemaDiff, true)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			w.walk(propertyFullName(propertyPath, fmt.Sprintf("anyOf[%s]", v)), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			w.walk(propertyFullName(propertyPath, fmt.Sprintf("oneOf[%s]", v)), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		w.walk(propertyFullName(propertyPath, "items"), "", schemaDiff.ItemsDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.PropertiesDiff != nil {
		if w.properties != nil {
			w.properties(propertyPath, schemaDiff, underAllOf, w.sharedAncestor)
		}
		for _, name := range slices.Sorted(maps.Keys(schemaDiff.PropertiesDiff.Modified)) {
			v := schemaDiff.PropertiesDiff.Modified[name]
			w.walk(propertyPath, name, v, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.AdditionalPropertiesDiff != nil {
		w.walk(propertyFullName(propertyPath, "additionalProperties"), "", schemaDiff.AdditionalPropertiesDiff, schemaDiff, underAllOf)
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 sub-schema fields
	if schemaDiff.PrefixItemsDiff != nil {
		for _, v := range schemaDiff.PrefixItemsDiff.Modified {
			w.walk(fmt.Sprintf("%s/prefixItems[%s]", propertyPath, v), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.ContainsDiff != nil {
		w.walk(fmt.Sprintf("%s/contains", propertyPath), "", schemaDiff.ContainsDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.PropertyNamesDiff != nil {
		w.walk(fmt.Sprintf("%s/propertyNames", propertyPath), "", schemaDiff.PropertyNamesDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.UnevaluatedItemsDiff != nil {
		w.walk(fmt.Sprintf("%s/unevaluatedItems", propertyPath), "", schemaDiff.UnevaluatedItemsDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.UnevaluatedPropertiesDiff != nil {
		w.walk(fmt.Sprintf("%s/unevaluatedProperties", propertyPath), "", schemaDiff.UnevaluatedPropertiesDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.IfDiff != nil {
		w.walk(fmt.Sprintf("%s/if", propertyPath), "", schemaDiff.IfDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.ThenDiff != nil {
		w.walk(fmt.Sprintf("%s/then", propertyPath), "", schemaDiff.ThenDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.ElseDiff != nil {
		w.walk(fmt.Sprintf("%s/else", propertyPath), "", schemaDiff.ElseDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.NotDiff != nil {
		w.walk(fmt.Sprintf("%s/not", propertyPath), "", schemaDiff.NotDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.ContentSchemaDiff != nil {
		w.walk(fmt.Sprintf("%s/contentSchema", propertyPath), "", schemaDiff.ContentSchemaDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.PatternPropertiesDiff != nil {
		for _, i := range slices.Sorted(maps.Keys(schemaDiff.PatternPropertiesDiff.Modified)) {
			v := schemaDiff.PatternPropertiesDiff.Modified[i]
			w.walk(fmt.Sprintf("%s/patternProperties[%s]", propertyPath, i), "", v, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.DependentSchemasDiff != nil {
		for _, i := range slices.Sorted(maps.Keys(schemaDiff.DependentSchemasDiff.Modified)) {
			v := schemaDiff.DependentSchemasDiff.Modified[i]
			w.walk(fmt.Sprintf("%s/dependentSchemas[%s]", propertyPath, i), "", v, schemaDiff, underAllOf)
		}
	}
}

func checkModifiedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	subschemaWalk{enter: func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff, _ bool, _ *sharedReach) {
		processor(propertyPath, propertyName, propertyItem, propertyParentItem)
	}}.walk("", "", schemaDiff, nil, false)
}

func checkAddedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff, underAllOf bool, shared *SharedSchema)) {
	if schemaDiff == nil {
		return
	}

	subschemaWalk{
		shared: sharedSchemas(schemaDiff, false),
		properties: func(propertyPath string, sd *diff.SchemaDiff, underAllOf bool, shared *sharedReach) {
			for _, name := range sd.PropertiesDiff.Added {
				processor(propertyPath, name, sd.Revision.Properties[name].Value, sd, underAllOf, shared.at(propertyFullName(propertyPath, name)))
			}
		}}.walk("", "", schemaDiff, nil, false)
}

func checkDeletedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff, underAllOf bool, shared *SharedSchema)) {
	if schemaDiff == nil {
		return
	}

	subschemaWalk{
		shared: sharedSchemas(schemaDiff, false),
		properties: func(propertyPath string, sd *diff.SchemaDiff, underAllOf bool, shared *sharedReach) {
			for _, name := range sd.PropertiesDiff.Deleted {
				processor(propertyPath, name, sd.Base.Properties[name].Value, sd, underAllOf, shared.at(propertyFullName(propertyPath, name)))
			}
		}}.walk("", "", schemaDiff, nil, false)
}
