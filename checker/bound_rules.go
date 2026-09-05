package checker

import (
	"sync"

	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/oasdiff/oasdiff/diff"
)

type boundSpec struct {
	idName  string // id segment, e.g. "max-length"
	keyword string // schema field name in claims and messages, e.g. "maxLength"
}

var boundSpecs = []boundSpec{
	{"max", "maximum"},
	{"min", "minimum"},
	{"multiple-of", "multipleOf"},
	{"max-length", "maxLength"},
	{"min-length", "minLength"},
	{"max-items", "maxItems"},
	{"min-items", "minItems"},
	{"max-properties", "maxProperties"},
	{"min-properties", "minProperties"},
	{"min-contains", "minContains"},
	{"max-contains", "maxContains"},
}

// schemaBound resolves a keyword to its diff.SchemaBound
func schemaBound(keyword string) (diff.SchemaBound, bool) {
	for _, bound := range diff.SchemaBounds {
		if bound.Keyword == keyword {
			return bound, true
		}
	}
	return diff.SchemaBound{}, false
}

// boundActions are the edits the generated rules cover. Setting a
// constraint narrows what the schema accepts and unsetting it widens;
// increase and decrease stay with the hand-written checks. An action with
// comment carries the shared explanatory comment on the cells where its
// verdict is breaking.
type boundAction struct {
	action string
	effect Effect
	// comment is the id of the shared comment explaining the action's verdict
	// where it is breaking; empty when the message speaks alone
	comment string
}

// boundSetComment explains the conservative verdict of setting a bound; the
// generated and hand-written set checks share it.
var boundSetComment = commentId("bound-set")

var (
	boundSet   = boundAction{"set", rules.EffectNarrows, boundSetComment}
	boundUnset = boundAction{"unset", rules.EffectWidens, ""}
)

var boundActions = []boundAction{boundSet, boundUnset}

// boundRuleComment returns the action's comment where the derived verdict is
// breaking, empty otherwise.
func boundRuleComment(action boundAction, level Level) string {
	if level == ERR {
		return action.comment
	}
	return ""
}

func directionName(direction Direction) string {
	if direction == DirectionResponse {
		return "response"
	}
	return "request"
}

func boundRuleId(direction Direction, scope, idName, action string) string {
	return directionName(direction) + "-" + scope + "-" + idName + "-" + action
}

// boundScopes are the schema roots the generated rules cover on each side:
// parameters exist on requests only, headers on responses only, and body and
// property share the media-type schema location.
func boundScopes(direction Direction) []string {
	if direction == DirectionResponse {
		return []string{"body", "property", "header"}
	}
	return []string{"body", "property", "parameter"}
}

func boundLocation(direction Direction, scope, keyword string) string {
	switch scope {
	case "parameter":
		return "paths.*.*.parameters.*.schema." + keyword
	case "header":
		return "paths.*.*.responses.*.headers.*.schema." + keyword
	case "body", "property":
		if direction == DirectionResponse {
			return "paths.*.*.responses.*.content.*.schema." + keyword
		}
		return "paths.*.*.requestBody.content.*.schema." + keyword
	}
	panic("unknown bound scope: " + scope)
}

func boundClaim(direction Direction, scope, keyword, action string) string {
	return boundLocation(direction, scope, keyword) + ":" + action
}

var (
	boundRulesOnce  sync.Once
	boundRulesList  BackwardCompatibilityRules
	handWrittenById map[string]bool
)

// boundRules generates the set/unset rules for every keyword in
// boundSpecs, one per direction, scope, and action, skipping the cells a
// hand-written rule already covers.
func boundRules() BackwardCompatibilityRules {
	boundRulesOnce.Do(func() {
		handWrittenById = map[string]bool{}
		for _, rule := range handWrittenRules() {
			handWrittenById[rule.Id] = true
		}
		for _, spec := range boundSpecs {
			for _, direction := range []Direction{DirectionRequest, DirectionResponse} {
				for _, scope := range boundScopes(direction) {
					for _, action := range boundActions {
						id := boundRuleId(direction, scope, spec.idName, action.action)
						if handWrittenById[id] {
							continue
						}
						boundRulesList = append(boundRulesList, newBackwardCompatibilityRule(
							id,
							rules.DeriveLevel(action.effect, direction),
							BoundSetUnsetCheck,
							direction,
							AreaSchema,
							KindConstraints,
							action.effect,
							nil,
							boundClaim(direction, scope, spec.keyword, action.action),
						))
					}
				}
			}
		}
	})
	return boundRulesList
}

// handWrittenIds reports the ids registered by hand-written rules, whose
// cells the generated check leaves to their own checks
func handWrittenIds() map[string]bool {
	boundRules()
	return handWrittenById
}

// classifySetUnset reports whether the keyword was set or unset, and the
// value that appeared or disappeared
func classifySetUnset(spec boundSpec, d *diff.SchemaDiff) (boundAction, any, bool) {
	bound, ok := schemaBound(spec.keyword)
	if !ok {
		return boundAction{}, nil, false
	}
	if value, ok := bound.WasSet(d); ok {
		return boundSet, value, true
	}
	if value, ok := bound.WasUnset(d); ok {
		return boundUnset, value, true
	}
	return boundAction{}, nil, false
}

// BoundSetUnsetCheck reports the set and unset changes for every keyword in
// boundSpecs, at body, property, parameter, and response-header level.
// Parameter and header root schemas attach no guards: readOnly and writeOnly
// are property-scoped, so there they declare nothing.
func BoundSetUnsetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, boundChanges(info, DirectionRequest, operationsSources)...)
	})
	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, boundChanges(info, DirectionResponse, operationsSources)...)
	})
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		result = append(result, boundSchemaChanges(p.paramDiff.SchemaDiff, DirectionRequest, "parameter", operationsSources, p.opInfo.methodDiff,
			func(value any) []any { return []any{p.location, p.name, value} },
			p.opInfo.NewApiChange)...)
	})
	walkModifiedResponseHeaders(diffReport, operationsSources, config, func(h headerInfo) {
		result = append(result, boundSchemaChanges(h.headerDiff.SchemaDiff, DirectionResponse, "header", operationsSources, h.opInfo.methodDiff,
			func(value any) []any { return []any{h.name, value, h.responseStatus} },
			h.opInfo.NewApiChange)...)
	})

	return result
}

// boundChanges reports the body-level and property-level cells of one media
// type; properties go through p.newChange, so the read-only and write-only
// guards attach as for every property check.
func boundChanges(info mediaTypeInfo, direction Direction, operationsSources *diff.OperationsSourcesMap) Changes {
	result := boundSchemaChanges(info.schemaDiff, direction, "body", operationsSources, info.operationItem,
		func(value any) []any { return []any{value} },
		info.newChange)

	info.walkProperties(func(p propertyInfo) {
		result = append(result, boundSchemaChanges(p.propertyDiff, direction, "property", operationsSources, info.operationItem,
			func(value any) []any {
				args := []any{propertyFullName(p.propertyPath, p.propertyName), value}
				if direction == DirectionResponse {
					args = append(args, info.responseStatus)
				}
				return args
			},
			p.newChange)...)
	})

	return result
}

// boundSchemaChanges reports the set and unset changes of one schema node:
// classify each keyword, skip the cells a hand-written check owns, and emit
// through the caller's change constructor with the caller's argument shape.
func boundSchemaChanges(
	schemaDiff *diff.SchemaDiff,
	direction Direction,
	scope string,
	operationsSources *diff.OperationsSourcesMap,
	methodDiff *diff.MethodDiff,
	args func(value any) []any,
	newChange func(id string, args []any, comment string) ApiChange,
) Changes {
	result := make(Changes, 0)
	if schemaDiff == nil {
		return result
	}
	for _, spec := range boundSpecs {
		action, value, ok := classifySetUnset(spec, schemaDiff)
		if !ok {
			continue
		}
		id := boundRuleId(direction, scope, spec.idName, action.action)
		if handWrittenIds()[id] {
			continue
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, methodDiff, schemaDiff, spec.keyword)
		result = append(result, newChange(
			id,
			args(value),
			boundRuleComment(action, rules.DeriveLevel(action.effect, direction)),
		).WithSources(baseSource, revisionSource))
	}
	return result
}
