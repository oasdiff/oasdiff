package internal

import "github.com/oasdiff/oasdiff/checker"

func getAllTags() []string {
	return []string{
		// direction
		"request", "response",
		// action
		"add", "remove", "change", "generalize", "specialize", "increase", "decrease", "set",
		// area (OpenAPI object)
		"schema", "parameters", "requestBody", "responses", "paths", "headers", "security", "tags", "components",
		// kind (aspect of the contract)
		"existence", "requiredness", "type", "constraints", "values", "structure", "lifecycle",
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

	if matchActionTag(tag, rule.Action) {
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

func matchActionTag(tag string, action checker.Action) bool {
	switch tag {
	case "add":
		return action == checker.ActionAdd
	case "remove":
		return action == checker.ActionRemove
	case "change":
		return action == checker.ActionChange
	case "generalize":
		return action == checker.ActionGeneralize
	case "specialize":
		return action == checker.ActionSpecialize
	case "increase":
		return action == checker.ActionIncrease
	case "decrease":
		return action == checker.ActionDecrease
	case "set":
		return action == checker.ActionSet
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
