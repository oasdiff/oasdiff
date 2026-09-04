package checker

import (
	"sync"

	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/oasdiff/oasdiff/diff"
)

// boundSpec names one ordered-constraint keyword: its id segment and the
// schema field name used in claims and messages. The keyword resolves to a
// diff.SchemaBound, which owns where the diff lives and how absence is
// encoded; the set/unset rules for every keyword are generated from this
// table (see boundRules), so a new ordered keyword is a row here, not a
// hand-written check per direction and scope.
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

// schemaBound resolves a keyword to its diff.SchemaBound; a keyword the diff
// does not list fails the behavioral gate, never emits silently.
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
// increase and decrease stay with the hand-written checks.
var boundActions = []struct {
	action string
	effect Effect
}{
	{"set", rules.EffectNarrows},
	{"unset", rules.EffectWidens},
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

func boundClaim(direction Direction, claimName, action string) string {
	if direction == DirectionResponse {
		return "paths.*.*.responses.*.content.*.schema." + claimName + ":" + action
	}
	return "paths.*.*.requestBody.content.*.schema." + claimName + ":" + action
}

// boundRules generates the set/unset rules for every keyword in
// boundSpecs, one per direction, scope, and action, skipping the cells a
// hand-written rule already covers. Each rule's level is derived with the
// severity law, its id follows the id grammar, and its handler is the shared
// BoundSetUnsetCheck.
var (
	boundRulesOnce   sync.Once
	boundRulesList   BackwardCompatibilityRules
	boundRuleIdsOnce map[string]bool
)

func boundRules() BackwardCompatibilityRules {
	boundRulesOnce.Do(func() {
		handWritten := map[string]bool{}
		for _, rule := range handWrittenRules() {
			handWritten[rule.Id] = true
		}
		boundRuleIdsOnce = map[string]bool{}
		for _, spec := range boundSpecs {
			for _, direction := range []Direction{DirectionRequest, DirectionResponse} {
				for _, scope := range []string{"body", "property"} {
					for _, action := range boundActions {
						id := boundRuleId(direction, scope, spec.idName, action.action)
						if handWritten[id] {
							continue
						}
						boundRuleIdsOnce[id] = true
						boundRulesList = append(boundRulesList, newBackwardCompatibilityRule(
							id,
							rules.DeriveLevel(action.effect, direction),
							BoundSetUnsetCheck,
							direction,
							AreaSchema,
							KindConstraints,
							action.effect,
							nil,
							boundClaim(direction, spec.keyword, action.action),
						))
					}
				}
			}
		}
	})
	return boundRulesList
}

// generatedBoundIds reports the ids boundRules generated;
// BoundSetUnsetCheck emits only these, leaving the hand-written cells to
// their own checks.
func generatedBoundIds() map[string]bool {
	boundRules()
	return boundRuleIdsOnce
}

// classifySetUnset reports whether the keyword was set or unset, and the
// value that appeared or disappeared. A change between two present values is
// an increase or decrease, which the hand-written checks own.
func classifySetUnset(spec boundSpec, d *diff.SchemaDiff) (string, any, bool) {
	bound, ok := schemaBound(spec.keyword)
	if !ok {
		return "", nil, false
	}
	if value, ok := bound.Set(d); ok {
		return "set", value, true
	}
	if value, ok := bound.Unset(d); ok {
		return "unset", value, true
	}
	return "", nil, false
}

// boundComment mirrors the hand-written convention: a request-side set is
// the conservative error verdict, so it carries the shared explanatory
// comment; the other cells speak through the message alone.
func boundComment(direction Direction, action, id string) string {
	if direction == DirectionRequest && action == "set" {
		return commentId(id)
	}
	return ""
}

// BoundSetUnsetCheck reports the set and unset changes for every keyword in
// boundSpecs, at body and property level, on both sides of the wire.
func BoundSetUnsetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, boundChanges(info, DirectionRequest, operationsSources)...)
	})
	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, boundChanges(info, DirectionResponse, operationsSources)...)
	})

	return result
}

func boundChanges(info mediaTypeInfo, direction Direction, operationsSources *diff.OperationsSourcesMap) Changes {
	result := make(Changes, 0)

	for _, spec := range boundSpecs {
		action, value, ok := classifySetUnset(spec, info.schemaDiff)
		if !ok {
			continue
		}
		id := boundRuleId(direction, "body", spec.idName, action)
		if !generatedBoundIds()[id] {
			continue
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, spec.keyword)
		result = append(result, info.newChange(
			id,
			[]any{value},
			boundComment(direction, action, id),
		).WithSources(baseSource, revisionSource))
	}

	info.walkProperties(func(p propertyInfo) {
		for _, spec := range boundSpecs {
			action, value, ok := classifySetUnset(spec, p.propertyDiff)
			if !ok {
				continue
			}
			id := boundRuleId(direction, "property", spec.idName, action)
			if !generatedBoundIds()[id] {
				continue
			}
			args := []any{propertyFullName(p.propertyPath, p.propertyName), value}
			if direction == DirectionResponse {
				args = append(args, info.responseStatus)
			}
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, spec.keyword)
			result = append(result, p.newChange(
				id,
				args,
				boundComment(direction, action, id),
			).WithSources(baseSource, revisionSource))
		}
	})

	return result
}
