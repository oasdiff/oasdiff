package checker

// CoverageWaiver explains one family of wire-relevant edits that no rule
// covers. Facts about the object model itself (fields outside the wire
// contract) live in metaschema.NonContracts; entries here are relative to
// the rule registry and go stale as checks are added.
type CoverageWaiver struct {
	// Pattern is a location glob (see metaschema.MatchLocation), optionally
	// restricted to actions with ":action[,action...]"; without the suffix
	// it waives every action at the location.
	Pattern string
	// Reason starts with a category: resolved-at-usage (component
	// definitions are compared at their referencing operations after $ref
	// resolution), covered-as (the same document edit is already reported
	// under another action at the same location), or open (a candidate
	// missing check, with its tracking issue).
	Reason string
}

// CoverageWaivers records why each wire-relevant edit with no rule is
// deliberately or knowingly uncovered. The checker's TestRuleCoverage fails
// on any uncovered edit no waiver matches (add a rule or a waiver) and on
// any waiver matching no uncovered edit (remove the stale waiver), so the
// list stays an honest, reviewed record.
var CoverageWaivers = []CoverageWaiver{
	// security schemes are consumed by name, never resolved through $ref
	// into usage sites, so the resolved-at-usage reasoning below does not
	// apply to them; this entry must precede components.**
	{"components.securitySchemes.**", "open: scheme fields beyond type, flow URLs, and scopes (apiKey name and in, http scheme, bearerFormat, openIdConnectUrl) have no checks (tracked in #1175)"},
	{"components.**", "resolved-at-usage: edits to component definitions surface as diffs at every referencing operation; only unused-component removal is reported directly (api-schemas-removed)"},
	{"webhooks.**", "open: webhooks are diffed (WebhooksDiff) but checkers only report webhook add/remove; changes inside a webhook's operations have no checks yet (tracked in #1160)"},
	{"paths.*.parameters.**", "open: path-level parameter additions are checked (new-request-*-default-parameter-to-existing-path); modifications and removals at path level have no checks yet (tracked in #1163)"},
	{"paths.*.*.callbacks.**", "open: callbacks are not checked (tracked in #1161)"},
	{"paths.*.*.requestBody.content.*.encoding.**", "open: multipart/form encoding metadata (contentType, per-part headers, style) has no checks (tracked in #1165)"},
	{"paths.*.*.responses.*.content.*.encoding.**", "open: encoding metadata has no checks (tracked in #1165)"},
	{"paths.*.*.parameters.*.content.**", "open: parameters serialized via a content map are not checked (only the schema form is) (tracked in #1166)"},
	{"paths.*.*.requestBody.content.*.itemSchema.**", "open: only itemSchema existence is checked (request-body-media-type-item-schema-added/removed); changes inside it are not (tracked in #1167)"},
	{"paths.*.*.responses.*.content.*.itemSchema.**", "open: only itemSchema existence is checked; changes inside it are not (tracked in #1167)"},
	{"paths.*.*.responses.*.headers.*.**", "open: response headers are checked for existence, required, and schema type/format/nullable only; serialization fields, the content form, and the remaining schema keywords are unchecked (tracked in #1162)"},
	{"paths.*.*.parameters.*.schema:set,unset", "open: a parameter schema appearing or disappearing is unchecked (the media-type analog has request-body-media-type-schema-added/removed) (tracked in #1054)"},
	{"paths.*.*.parameters.*.schema.**", "open: parameter schemas are checked for type/format, enum, bounds, pattern, nullable, default, and required/property membership; the remaining schema keywords are unchecked (tracked in #1054, #1155, #1156, #1157, #1159)"},
	{"paths.*.*.parameters.*.style", "open: parameter serialization style changes the wire format but is unchecked (tracked in #1164)"},
	{"paths.*.*.parameters.*.explode", "open: explode changes the wire format of array/object parameters but is unchecked (tracked in #1164)"},
	{"paths.*.*.parameters.*.allowReserved", "open: allowReserved changes accepted query characters but is unchecked (tracked in #1164)"},
	{"**.discriminator.mapping.*:set,unset", "covered-as add/remove: a mapping entry appearing or disappearing is the entry add/remove, which is claimed"},
	{"**.discriminator.propertyName:set,unset", "covered-as discriminator set/unset: propertyName is required inside discriminator, so its presence tracks the discriminator's"},
	{"**.schema.additionalProperties", "open: setting additionalProperties:false narrows accepted request objects (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.multipleOf", "open: response set (narrowing the output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.uniqueItems", "open: response set (narrowing the output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.maxProperties", "open: remaining directions and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minProperties", "open: remaining directions and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159, #1171 for the set case)"},
	{"**.schema.items:set,unset", "open: an items subschema appearing on a request narrows accepted arrays (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.not", "open: a not subschema appearing on a request narrows the accepted set (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.maxItems", "open: remaining directions (request unset widens, response set/decrease narrow the server's output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.maximum", "open: remaining directions (request unset widens, response set/decrease narrows the server's output) are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minimum", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.maxLength", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minLength", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minItems", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.exclusiveMaximum", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.exclusiveMinimum", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minContains:set,unset", "open: minContains presence changes are unchecked; increase/decrease have rules (tracked in #1159)"},
	{"**.schema.maxContains:set,unset", "open: maxContains presence changes are unchecked; increase/decrease have rules (tracked in #1159)"},
	{"**.schema.unevaluatedItems:change", "open: switching unevaluatedItems between boolean and schema form is unchecked; set/unset have rules (tracked in #1054)"},
	{"**.schema.unevaluatedProperties:change", "open: switching unevaluatedProperties between boolean and schema form is unchecked; set/unset have rules (tracked in #1054)"},
}
