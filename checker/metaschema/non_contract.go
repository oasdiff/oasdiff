package metaschema

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
	// The only $ref fields that appear as edits are path items' (in paths,
	// webhooks, and callbacks): every other ref is a wrapper type the walk
	// dissolves into its resolved value.
	{"**.$ref", "a $ref is resolved at load time; the diff compares resolved content"},
	{"**.servers.**", "server URLs are deployment metadata, not part of the request/response contract"},
	{"**.links.**", "links document relationships between operations; they do not change accepted payloads"},
	{"openapi", "the OpenAPI dialect version governs parsing, not the API contract"},
	{"jsonSchemaDialect", "schema dialect governs parsing, not the API contract"},
	{"info", "the info object is required metadata; its presence is a validation concern"},
	{"info.version:set,unset", "version is required; presence changes are validation errors, value changes are checked (api-version rules)"},
	{"**.parameters.*.name", "parameters are identified by (name, in); a rename surfaces as parameter remove + add"},
	{"**.parameters.*.in", "parameters are identified by (name, in); a location change surfaces as parameter remove + add"},
	// headers share the parameter type, but the spec forbids name and in
	// on Header objects: a header is identified by its map key
	{"**.headers.*.name", "a header is identified by its map key; the name field is forbidden on Header objects"},
	{"**.headers.*.in", "a header is identified by its map key; the in field is forbidden on Header objects"},
	{"**.allowEmptyValue", "allowEmptyValue is deprecated by the OpenAPI spec, which advises against its use"},
	{"**.schema.$id", "reference-resolution keyword; its effect is visible in the resolved schemas the diff compares"},
	{"**.schema.$anchor", "reference-resolution keyword"},
	{"**.schema.$dynamicAnchor", "reference-resolution keyword"},
	{"**.schema.$dynamicRef", "reference-resolution keyword"},
	{"**.schema.$schema", "schema dialect declaration, governs validation semantics at parse time"},
	{"**.schema.$defs.**", "$defs holds definitions that take effect only where referenced"},
	{"**.schema.$comment", "comments carry no validation semantics"},
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
