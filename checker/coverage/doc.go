// oasdiff's ultimate objective is to detect breaking changes, so that its
// users can prevent them. The diff layer reports every change to a spec, but
// falls short of that objective in two ways: it does not classify changes as
// breaking or not, and it reports them in the spec's own hierarchical,
// technical vocabulary rather than as human-readable descriptions. The
// checker closes both gaps with rules that detect kinds of changes, judge
// them, and report them in user-facing language.
//
// The rules were defined gradually, each covering one more kind of change,
// which leaves the central question unanswerable: how well does oasdiff do
// its job? Which changes does it not detect? Auditing the rules against each
// other cannot answer it, because consistency cannot reveal what is absent
// altogether.
//
// Package coverage computes the answer. Analyze measures the rules against
// every possible edit of an OpenAPI document (see checker/metaschema) and
// decides each edit: which checks cover it, or why none are expected. Each
// rule declares the edits it covers, and every edit that can change which
// payloads are valid must be covered by a check or waived with a reviewed
// reason. An edit that is neither is a gap; a waiver that accounts for
// nothing is stale; and an unchecked edit that a check could cover carries a
// suggested id for it. Completeness is thereby measured rather than assumed.
//
// Analyze takes the rules to measure as an argument rather than calling
// checker.GetAllRules itself, so this package needs only the rule model and
// the enumeration, never the code that runs the checks.
package coverage
