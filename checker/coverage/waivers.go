package coverage

// WaiverCategory says what kind of accounting a waiver is, which decides
// whether a missing check is implied: only open waivers describe checks
// that could exist.
type WaiverCategory string

const (
	// CategoryOpen: a candidate missing check, with its tracking issue.
	CategoryOpen WaiverCategory = "open"
	// CategoryResolvedAtUsage: component definitions are compared at their
	// referencing operations after $ref resolution, so the check belongs at
	// the usage sites, which have their own edits.
	CategoryResolvedAtUsage WaiverCategory = "resolved-at-usage"
	// CategoryCoveredAs: the same document edit is already reported under
	// another action at the same location.
	CategoryCoveredAs WaiverCategory = "covered-as"
)

// Waiver explains one family of wire-relevant edits that no rule
// covers. Facts about the object model itself (fields outside the wire
// contract) live in metaschema.NonContracts; entries here describe what the
// checks happen to miss today, and go stale as checks are added.
type Waiver struct {
	Category WaiverCategory
	// Pattern is a location glob (see metaschema.MatchLocation), optionally
	// restricted to actions with ":action[,action...]"; without the suffix
	// it waives every action at the location.
	Pattern string
	Reason  string
}

// Waivers records why each wire-relevant edit with no rule is
// deliberately or knowingly uncovered. An uncovered edit no waiver matches
// (add a rule or a waiver) and a waiver matching no uncovered edit (remove
// the stale entry) both fail the build, so the list stays an honest,
// reviewed record.
var Waivers = []Waiver{
	// security schemes are consumed by name, never resolved through $ref
	// into usage sites, so the resolved-at-usage reasoning below does not
	// apply to them; this entry must precede components.**
	{CategoryOpen, "components.securitySchemes.**", "scheme fields beyond type, flow URLs, and scopes (apiKey name and in, http scheme, bearerFormat, openIdConnectUrl) have no checks (tracked in #1175)"},
	{CategoryResolvedAtUsage, "components.**", "edits to component definitions surface as diffs at every referencing operation; only unused-component removal is reported directly (api-schemas-removed)"},
	{CategoryOpen, "webhooks.**", "webhooks are diffed (WebhooksDiff) but checkers only report webhook add/remove; changes inside a webhook's operations have no checks yet (tracked in #1160)"},
	{CategoryOpen, "paths.*.parameters.**", "path-level parameter additions are checked (new-request-*-default-parameter-to-existing-path); modifications and removals at path level have no checks yet (tracked in #1163)"},
	{CategoryOpen, "paths.*.*.callbacks.**", "callbacks are not checked (tracked in #1161)"},
	{CategoryOpen, "paths.*.*.requestBody.content.*.encoding.**", "multipart/form encoding metadata (contentType, per-part headers, style) has no checks (tracked in #1165)"},
	{CategoryOpen, "paths.*.*.responses.*.content.*.encoding.**", "encoding metadata has no checks (tracked in #1165)"},
	{CategoryOpen, "paths.*.*.parameters.*.content.**", "parameters serialized via a content map are not checked (only the schema form is) (tracked in #1166)"},
	{CategoryOpen, "paths.*.*.requestBody.content.*.itemSchema.**", "only itemSchema existence is checked (request-body-media-type-item-schema-added/removed); changes inside it are not (tracked in #1167)"},
	{CategoryOpen, "paths.*.*.responses.*.content.*.itemSchema.**", "only itemSchema existence is checked; changes inside it are not (tracked in #1167)"},
	{CategoryOpen, "paths.*.*.responses.*.headers.*.**", "response headers are checked for existence, required, and schema type/format/nullable only; serialization fields, the content form, and the remaining schema keywords are unchecked (tracked in #1162)"},
	{CategoryOpen, "paths.*.*.parameters.*.schema:set,unset", "a parameter schema appearing or disappearing is unchecked (the media-type analog has request-body-media-type-schema-added/removed) (tracked in #1054)"},
	{CategoryOpen, "paths.*.*.parameters.*.schema.**", "parameter schemas are checked for type/format, enum, bounds, pattern, nullable, default, and required/property membership; the remaining schema keywords are unchecked (tracked in #1054, #1155, #1156, #1157, #1159)"},
	{CategoryOpen, "paths.*.*.parameters.*.style", "parameter serialization style changes the wire format but is unchecked (tracked in #1164)"},
	{CategoryOpen, "paths.*.*.parameters.*.explode", "explode changes the wire format of array/object parameters but is unchecked (tracked in #1164)"},
	{CategoryOpen, "paths.*.*.parameters.*.allowReserved", "allowReserved changes accepted query characters but is unchecked (tracked in #1164)"},
	{CategoryCoveredAs, "**.discriminator.mapping.*:set,unset", "add/remove: a mapping entry appearing or disappearing is the entry add/remove, which is claimed"},
	{CategoryCoveredAs, "**.discriminator.propertyName:set,unset", "discriminator set/unset: propertyName is required inside discriminator, so its presence tracks the discriminator's"},
	{CategoryOpen, "**.schema.additionalProperties", "setting additionalProperties:false narrows accepted request objects (breaking) and is unchecked (tracked in #1054)"},
	{CategoryOpen, "**.schema.multipleOf", "response set (narrowing the output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.uniqueItems", "response set (narrowing the output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.maxProperties", "remaining directions and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.minProperties", "remaining directions and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159, #1171 for the set case)"},
	{CategoryOpen, "**.schema.items:set,unset", "an items subschema appearing on a request narrows accepted arrays (breaking) and is unchecked (tracked in #1054)"},
	{CategoryOpen, "**.schema.not", "a not subschema appearing on a request narrows the accepted set (breaking) and is unchecked (tracked in #1054)"},
	{CategoryOpen, "**.schema.maxItems", "remaining directions (request unset widens, response set/decrease narrow the server's output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.maximum", "remaining directions (request unset widens, response set/decrease narrows the server's output) are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.minimum", "remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.maxLength", "remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.minLength", "remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.minItems", "remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.exclusiveMaximum", "remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.exclusiveMinimum", "remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.minContains:set,unset", "minContains presence changes are unchecked; increase/decrease have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.maxContains:set,unset", "maxContains presence changes are unchecked; increase/decrease have rules (tracked in #1159)"},
	{CategoryOpen, "**.schema.unevaluatedItems:change", "switching unevaluatedItems between boolean and schema form is unchecked; set/unset have rules (tracked in #1054)"},
	{CategoryOpen, "**.schema.unevaluatedProperties:change", "switching unevaluatedProperties between boolean and schema form is unchecked; set/unset have rules (tracked in #1054)"},
}
