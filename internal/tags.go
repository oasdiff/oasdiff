package internal

import (
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/metaschema"
)

func getAllTags() []string {
	return []string{
		// direction
		"request", "response",
		// action (syntactic edit, from the rule's location claims)
		"add", "remove", "change", "increase", "decrease", "set", "unset",
		// effect (semantic verdict); generalize/specialize are aliases kept
		// from the retired action vocabulary
		"widens", "narrows", "generalize", "specialize",
		// area (OpenAPI object)
		"schema", "parameters", "requestBody", "responses", "paths", "headers", "security", "tags", "components",
		// kind (aspect of the contract)
		"existence", "requiredness", "mutability", "type", "constraints", "values", "structure", "lifecycle",
	}
}

// matchTags returns true if the rule matches all the tags
func matchTags(tags []string, rule checker.BackwardCompatibilityRule) bool {
	if len(tags) == 0 {
		return true
	}

	for _, tag := range tags {
		if !matchTag(tag, rule) {
			return false
		}
	}

	return true
}

func matchTag(tag string, rule checker.BackwardCompatibilityRule) bool {
	if matchAreaTag(tag, rule.Area) {
		return true
	}

	if matchKindTag(tag, rule.Kind) {
		return true
	}

	if matchActionTag(tag, rule.Actions()) {
		return true
	}

	if matchEffectTag(tag, rule.Effect) {
		return true
	}

	if matchDirectionTag(tag, rule.Direction) {
		return true
	}

	return false
}

func matchDirectionTag(tag string, direction checker.Direction) bool {
	switch tag {
	case "request":
		return direction == checker.DirectionRequest
	case "response":
		return direction == checker.DirectionResponse
	}

	return false
}

func matchActionTag(tag string, actions []metaschema.Action) bool {
	switch tag {
	case "add", "remove", "change", "increase", "decrease", "set", "unset":
		return slices.Contains(actions, metaschema.Action(tag))
	}

	return false
}

func matchEffectTag(tag string, effect checker.Effect) bool {
	switch tag {
	case "widens", "generalize":
		return effect == checker.EffectWidens
	case "narrows", "specialize":
		return effect == checker.EffectNarrows
	}

	return false
}

func matchAreaTag(tag string, area checker.Area) bool {
	switch tag {
	case "schema":
		return area == checker.AreaSchema
	case "parameters":
		return area == checker.AreaParameters
	case "requestBody":
		return area == checker.AreaRequestBody
	case "responses":
		return area == checker.AreaResponses
	case "paths":
		return area == checker.AreaPaths
	case "headers":
		return area == checker.AreaHeaders
	case "security":
		return area == checker.AreaSecurity
	case "tags":
		return area == checker.AreaTags
	case "components":
		return area == checker.AreaComponents
	}

	return false
}

func matchKindTag(tag string, kind checker.Kind) bool {
	switch tag {
	case "existence":
		return kind == checker.KindExistence
	case "requiredness":
		return kind == checker.KindRequiredness
	case "mutability":
		return kind == checker.KindMutability
	case "type":
		return kind == checker.KindType
	case "constraints":
		return kind == checker.KindConstraints
	case "values":
		return kind == checker.KindValues
	case "structure":
		return kind == checker.KindStructure
	case "lifecycle":
		return kind == checker.KindLifecycle
	}

	return false
}

func joinActions(actions []metaschema.Action) string {
	strs := make([]string, len(actions))
	for i, a := range actions {
		strs[i] = string(a)
	}
	return strings.Join(strs, ",")
}
