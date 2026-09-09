package checker

import (
	"sync"

	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/oasdiff/oasdiff/diff"
)

type boundSpec struct {
	idName   string // id segment, e.g. "max-length"
	keyword  string // schema field name in claims and messages, e.g. "maxLength"
	polarity boundPolarity
}

// boundPolarity is which side of the value range a bound constrains, and so
// which of increasing or decreasing it narrows the accepted values.
type boundPolarity int

const (
	lowerBound boundPolarity = iota // increasing narrows
	upperBound                      // decreasing narrows
	// numeric order does not decide the effect (multipleOf narrows by
	// divisibility): no increase or decrease cells are generated
	unordered
)

var boundSpecs = []boundSpec{
	{"max", "maximum", upperBound},
	{"min", "minimum", lowerBound},
	{"multiple-of", "multipleOf", unordered},
	{"max-length", "maxLength", upperBound},
	{"min-length", "minLength", lowerBound},
	{"max-items", "maxItems", upperBound},
	{"min-items", "minItems", lowerBound},
	{"max-properties", "maxProperties", upperBound},
	{"min-properties", "minProperties", lowerBound},
	{"min-contains", "minContains", lowerBound},
	{"max-contains", "maxContains", upperBound},
	{"exclusive-min", "exclusiveMinimum", lowerBound},
	{"exclusive-max", "exclusiveMaximum", upperBound},
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

// Bounds the hand-written set checks classify through. A wrong keyword here
// leaves the bound zero and its check silent; the checks' fixture tests and
// TestBoundSpecsMatchSchemaBounds keep that loud.
var (
	maximumBound, _          = schemaBound("maximum")
	minimumBound, _          = schemaBound("minimum")
	exclusiveMaximumBound, _ = schemaBound("exclusiveMaximum")
	exclusiveMinimumBound, _ = schemaBound("exclusiveMinimum")
)

// boundActions are the edits the generated rules cover. An action with
// comment carries the shared explanatory comment on the cells where its
// verdict is breaking.
type boundAction struct {
	verb  string // id segment and message key, e.g. "increased"
	claim string // metaschema action in claims, e.g. "increase"
	// comment is the id of the shared comment explaining the action's verdict
	// where it is breaking; empty when the message speaks alone
	comment string
}

// boundSetComment explains the conservative verdict of setting a bound; the
// generated and hand-written set checks share it.
var boundSetComment = commentId("bound-set")

var (
	boundSet       = boundAction{"set", "set", boundSetComment}
	boundUnset     = boundAction{"unset", "unset", ""}
	boundIncreased = boundAction{"increased", "increase", ""}
	boundDecreased = boundAction{"decreased", "decrease", ""}
)

var boundActions = []boundAction{boundSet, boundUnset, boundIncreased, boundDecreased}

// boundEffect derives what the action does to the accepted values. Setting a
// bound narrows and unsetting it widens, whichever bound it is; increasing
// and decreasing depend on the bound's polarity, and have no cells on a
// keyword whose effect numeric order does not decide.
func boundEffect(polarity boundPolarity, action boundAction) (Effect, bool) {
	switch action {
	case boundSet:
		return rules.EffectNarrows, true
	case boundUnset:
		return rules.EffectWidens, true
	}
	if polarity == unordered {
		return rules.EffectNone, false
	}
	if (action == boundIncreased) == (polarity == lowerBound) {
		return rules.EffectNarrows, true
	}
	return rules.EffectWidens, true
}

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
	// an unknown scope yields an empty location, and the claims audit
	// rejects the malformed claim built from it
	return ""
}

func boundClaim(direction Direction, scope, keyword, action string) string {
	return boundLocation(direction, scope, keyword) + ":" + action
}

var (
	boundRulesOnce  sync.Once
	boundRulesList  BackwardCompatibilityRules
	handWrittenById map[string]bool
)

// boundRules generates the set, unset, increased, and decreased rules for
// every keyword in boundSpecs, one per direction, scope, and action, skipping
// the cells a hand-written rule already covers.
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
						effect, ok := boundEffect(spec.polarity, action)
						if !ok {
							continue
						}
						id := boundRuleId(direction, scope, spec.idName, action.verb)
						if handWrittenById[id] {
							continue
						}
						boundRulesList = append(boundRulesList, newBackwardCompatibilityRule(
							id,
							rules.DeriveLevel(effect, direction),
							BoundCheck,
							direction,
							AreaSchema,
							KindConstraints,
							effect,
							nil,
							boundClaim(direction, scope, spec.keyword, action.claim),
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

// classifyBound reports which bound action the keyword's diff is, with the
// message values: the appearing or disappearing value for set and unset, the
// from and to values for increased and decreased. Increase and decrease are
// only classified where numeric order decides the effect, mirroring the rule
// generation.
func classifyBound(spec boundSpec, d *diff.SchemaDiff) (boundAction, []any, bool) {
	bound, ok := schemaBound(spec.keyword)
	if !ok {
		return boundAction{}, nil, false
	}
	if value, ok := bound.WasSet(d); ok {
		return boundSet, []any{value}, true
	}
	if value, ok := bound.WasUnset(d); ok {
		return boundUnset, []any{value}, true
	}
	if spec.polarity == unordered {
		return boundAction{}, nil, false
	}
	if from, to, ok := bound.WasIncreased(d); ok {
		return boundIncreased, []any{from, to}, true
	}
	if from, to, ok := bound.WasDecreased(d); ok {
		return boundDecreased, []any{from, to}, true
	}
	return boundAction{}, nil, false
}

// BoundCheck reports the set, unset, increased, and decreased changes for
// every keyword in boundSpecs, at body, property, parameter, and
// response-header level. Parameter and header root schemas attach no guards:
// readOnly and writeOnly are property-scoped, so there they declare nothing.
func BoundCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, boundChanges(info, DirectionRequest, operationsSources)...)
	})
	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, boundChanges(info, DirectionResponse, operationsSources)...)
	})
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		result = append(result, boundSchemaChanges(p.paramDiff.SchemaDiff, DirectionRequest, "parameter", operationsSources, p.opInfo.methodDiff,
			func(values []any) []any { return append([]any{p.location, p.name}, values...) },
			p.opInfo.NewApiChange)...)
	})
	walkModifiedResponseHeaders(diffReport, operationsSources, config, func(h headerInfo) {
		result = append(result, boundSchemaChanges(h.headerDiff.SchemaDiff, DirectionResponse, "header", operationsSources, h.opInfo.methodDiff,
			func(values []any) []any { return append(append([]any{h.name}, values...), h.responseStatus) },
			h.opInfo.NewApiChange)...)
	})

	return result
}

// boundChanges reports the body-level and property-level cells of one media
// type; properties go through p.newChange, so the read-only and write-only
// guards attach as for every property check.
func boundChanges(info mediaTypeInfo, direction Direction, operationsSources *diff.OperationsSourcesMap) Changes {
	result := boundSchemaChanges(info.schemaDiff, direction, "body", operationsSources, info.operationItem,
		func(values []any) []any { return values },
		info.newChange)

	info.walkProperties(func(p propertyInfo) {
		result = append(result, boundSchemaChanges(p.propertyDiff, direction, "property", operationsSources, info.operationItem,
			func(values []any) []any {
				args := append([]any{propertyFullName(p.propertyPath, p.propertyName)}, values...)
				if direction == DirectionResponse {
					args = append(args, info.responseStatus)
				}
				return args
			},
			p.newChange)...)
	})

	return result
}

// boundSchemaChanges reports the bound changes of one schema node:
// classify each keyword, skip the cells a hand-written check owns, and emit
// through the caller's change constructor with the caller's argument shape.
func boundSchemaChanges(
	schemaDiff *diff.SchemaDiff,
	direction Direction,
	scope string,
	operationsSources *diff.OperationsSourcesMap,
	methodDiff *diff.MethodDiff,
	args func(values []any) []any,
	newChange func(id string, args []any, comment string) ApiChange,
) Changes {
	result := make(Changes, 0)
	if schemaDiff == nil {
		return result
	}
	for _, spec := range boundSpecs {
		action, values, ok := classifyBound(spec, schemaDiff)
		if !ok {
			continue
		}
		id := boundRuleId(direction, scope, spec.idName, action.verb)
		if handWrittenIds()[id] {
			continue
		}
		effect, _ := boundEffect(spec.polarity, action)
		baseSource, revisionSource := SchemaFieldSources(operationsSources, methodDiff, schemaDiff, spec.keyword)
		result = append(result, newChange(
			id,
			args(values),
			boundRuleComment(action, rules.DeriveLevel(effect, direction)),
		).WithSources(baseSource, revisionSource))
	}
	return result
}
