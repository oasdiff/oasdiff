// oasdiff's ultimate objective is to detect breaking changes, so that its
// users can prevent them. The diff layer reports every change to a spec,
// but falls short of that objective in two ways: it does not classify
// changes as breaking or not, and it reports them in the spec's own
// hierarchical, technical vocabulary rather than as human-readable
// descriptions. The checker closes both gaps with rules that detect kinds
// of changes, judge them, and report them in user-facing language.
//
// The rules were defined gradually, each covering one more kind of change,
// which leaves the central question unanswerable: how well does oasdiff do
// its job? Which changes does it not detect? The checker's symmetry and
// monotonicity tests audit the rules for consistency with each other, but
// consistency cannot reveal what is absent altogether.
//
// Package metaschema answers the question analytically: it enumerates
// every possible edit of an OpenAPI document — each field location in the
// OpenAPI object model, paired with the syntactic edit actions applicable
// at that location — so the rules can be audited against the full space of
// changes. Every rule declares which edits it covers, and every edit must
// be covered by a rule or carry a reviewed reason why not; the gap between
// the two is thereby identified, tracked, and closed. Completeness stops
// being a claim and becomes a test: the build fails while the accounting
// is incomplete, and `oasdiff checks changelog coverage` renders it.
//
// The enumeration is derived by reflection from the kin-openapi types
// oasdiff parses specs into, so a field added upstream shows up as new
// edits without a manual update here — and the audit fails until the rules
// account for them.
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
