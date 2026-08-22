package internal

import (
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// A tagDimension is one axis of a tag vocabulary: its values and how a value
// matches a row. Tag filtering is OR within a dimension and AND across
// dimensions: --tags request,response,add selects rows that are (request or
// response) and add. Values must be unique across the dimensions of one
// vocabulary so every tag names exactly one axis.
type tagDimension[T any] struct {
	values []string
	match  func(value string, row T) bool
}

func tagValues[T any](dimensions []tagDimension[T]) []string {
	var result []string
	for _, d := range dimensions {
		result = append(result, d.values...)
	}
	return result
}

func matchTagDimensions[T any](tags []string, dimensions []tagDimension[T], row T) bool {
	for _, d := range dimensions {
		matched, requested := false, false
		for _, tag := range tags {
			if !slices.Contains(d.values, tag) {
				continue
			}
			requested = true
			if d.match(tag, row) {
				matched = true
				break
			}
		}
		if requested && !matched {
			return false
		}
	}
	return true
}

// changelogTagDimensions is the tag vocabulary of `checks changelog`:
// direction, action (the syntactic edits from the rule's location claims),
// effect (the rule's verdict; generalize and specialize are aliases kept
// from the retired action vocabulary), area, and kind.
var changelogTagDimensions = []tagDimension[checker.BackwardCompatibilityRule]{
	{
		values: []string{"request", "response"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			switch value {
			case "request":
				return rule.Direction == checker.DirectionRequest
			case "response":
				return rule.Direction == checker.DirectionResponse
			}
			return false
		},
	},
	{
		values: []string{"add", "remove", "change", "increase", "decrease", "set", "unset"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return slices.Contains(rule.Actions(), metaschema.Action(value))
		},
	},
	{
		values: []string{"widens", "narrows", "generalize", "specialize"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			switch value {
			case "widens", "generalize":
				return rule.Effect == checker.EffectWidens
			case "narrows", "specialize":
				return rule.Effect == checker.EffectNarrows
			}
			return false
		},
	},
	{
		values: []string{"schema", "parameters", "requestBody", "responses", "paths", "headers", "security", "tags", "components"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return matchAreaTag(value, rule.Area)
		},
	},
	{
		values: []string{"existence", "requiredness", "mutability", "type", "constraints", "values", "structure", "lifecycle"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return matchKindTag(value, rule.Kind)
		},
	},
}

// coverageTagDimensions is the tag vocabulary of `checks changelog coverage`:
// the audit status, the edit's polarity, and the edit's action.
var coverageTagDimensions = []tagDimension[checker.CoverageEdit]{
	{
		values: []string{"covered", "uncovered", "waived", "non-contract"},
		match: func(value string, edit checker.CoverageEdit) bool {
			return value == string(edit.Status)
		},
	},
	{
		values: []string{"request", "response", "document", "shared"},
		match: func(value string, edit checker.CoverageEdit) bool {
			return value == edit.Polarity
		},
	},
	{
		values: []string{"add", "remove", "change", "increase", "decrease", "set", "unset"},
		match: func(value string, edit checker.CoverageEdit) bool {
			return value == edit.Action
		},
	},
}

func getChangelogTags() []string {
	return tagValues(changelogTagDimensions)
}

func getCoverageTags() []string {
	return tagValues(coverageTagDimensions)
}

func matchChangelogTags(tags []string, rule checker.BackwardCompatibilityRule) bool {
	return matchTagDimensions(tags, changelogTagDimensions, rule)
}

func matchCoverageTags(tags []string, edit checker.CoverageEdit) bool {
	return matchTagDimensions(tags, coverageTagDimensions, edit)
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
