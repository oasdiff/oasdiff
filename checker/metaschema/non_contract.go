package metaschema

import "strings"

// NonContract marks a family of possible edits as outside the wire
// contract: changing these fields never changes which payloads are valid,
// so the coverage audit does not expect checks for them. Unlike a coverage
// waiver in the checker, an entry here is a fact about the OpenAPI object
// model itself, independent of which checks exist.
type NonContract struct {
	// Pattern is a location glob (see MatchLocation), optionally restricted
	// to actions with ":action[,action...]"; without the suffix it covers
	// every action at the location.
	Pattern string
	Reason  string
}

var NonContracts = []NonContract{
	{"paths.*.$ref", "a path item $ref is resolved at load time; the diff compares resolved content"},
	{"servers.**", "server URLs are deployment metadata, not part of the request/response contract"},
	{"paths.*.servers.**", "server URLs are deployment metadata"},
	{"paths.*.*.servers.**", "server URLs are deployment metadata"},
	{"paths.*.*.responses.*.links.**", "links document relationships between operations; they do not change accepted payloads"},
	{"openapi", "the OpenAPI dialect version governs parsing, not the API contract"},
	{"jsonSchemaDialect", "schema dialect governs parsing, not the API contract"},
	{"info", "the info object is required metadata; its presence is a validation concern"},
	{"info.version:set,unset", "version is required; presence changes are validation errors, value changes are checked (api-version rules)"},
	{"paths.*.*.parameters.*.name", "parameters are identified by (name, in); a rename surfaces as parameter remove + add"},
	{"paths.*.*.parameters.*.in", "parameters are identified by (name, in); a location change surfaces as parameter remove + add"},
	{"paths.*.*.parameters.*.allowEmptyValue", "allowEmptyValue is deprecated by the OpenAPI spec, which advises against its use"},
	{"**.schema.$id", "reference-resolution keyword; its effect is visible in the resolved schemas the diff compares"},
	{"**.schema.$anchor", "reference-resolution keyword"},
	{"**.schema.$dynamicAnchor", "reference-resolution keyword"},
	{"**.schema.$dynamicRef", "reference-resolution keyword"},
	{"**.schema.$schema", "schema dialect declaration, governs validation semantics at parse time"},
	{"**.schema.$defs.**", "$defs holds definitions that take effect only where referenced"},
	{"**.schema.$comment", "comments carry no validation semantics"},
	{"**.schema.allowEmptyValue", "allowEmptyValue is a parameter-object field; on a schema it has no effect"},
}

// MatchPattern reports whether a waiver-style pattern (a location glob with
// an optional ":action[,action...]" restriction) covers the edit.
func MatchPattern(pattern string, edit Edit) (bool, error) {
	if !strings.Contains(pattern, ":") {
		return MatchLocation(pattern, edit.Location), nil
	}
	claim, err := ParseClaim(pattern)
	if err != nil {
		return false, err
	}
	return claim.Matches(edit), nil
}

// WireRelevant reports whether the edit can change which payloads are
// valid: not an annotation, not a specification extension, and not covered
// by a NonContract entry.
func WireRelevant(edit Edit) bool {
	if edit.Annotation || edit.Extension {
		return false
	}
	for _, nc := range NonContracts {
		if matches, _ := MatchPattern(nc.Pattern, edit); matches {
			return false
		}
	}
	return true
}
