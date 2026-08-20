// Package metaschema exists to make one question answerable by machine:
// does oasdiff account for every change that can be made to an OpenAPI
// document? A checker built rule by rule can only hope its rule set is
// complete; a missing check looks exactly like silence. The way out is to
// enumerate the space of possible changes independently of the rules, so
// the rules can be audited against it: every rule declares which edits it
// covers, and every edit must be covered by a rule or carry a reviewed
// reason why not. Completeness stops being a claim and becomes a test
// (the checker's TestRuleCoverage; `oasdiff checks changelog coverage`
// renders the same accounting).
//
// To that end, the package enumerates every possible edit of an OpenAPI
// document: each field location in the OpenAPI object model, paired with
// the syntactic edit actions applicable at that location. The enumeration
// is derived by reflection from the kin-openapi types oasdiff parses specs
// into, so a field added upstream shows up as new edits without a manual
// update here — and the audit fails until the rules account for them.
//
// Actions are syntactic (add, remove, set, unset, change, increase,
// decrease), not verdicts: whether an edit is breaking is the checker's
// judgment, recorded by mapping rules to edits. Polarity is the
// syntactic position in the document (request, response, shared, document);
// semantic direction inversions (callbacks, webhooks, not-schemas,
// readOnly/writeOnly) are likewise the checker's business.
//
// Schema is a recursive node: the walk stops when a type reappears on the
// current path, so an edit like paths.*.*.requestBody.content.*.schema.maxLength
// stands for that keyword at any nesting depth inside the request body schema.
package metaschema
