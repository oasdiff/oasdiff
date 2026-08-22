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
// its job? Which changes does it not detect? The checker's symmetry and
// severity-law tests audit the rules for consistency with each other, but
// consistency cannot reveal what is absent altogether.
//
// Package coverage computes the answer. Analyze measures the rules against
// every possible edit of an OpenAPI document (see checker/metaschema) and
// decides each edit: which checks cover it, or why none are expected. Each
// rule declares the edits it covers, and every edit that can change which
// payloads are valid must be covered by a check or waived with a reviewed
// reason.
//
// Two callers act on that: the checker's tests, which fail the build on an
// edit that is neither covered nor waived and on a waiver that no longer
// accounts for anything; and `oasdiff checks changelog coverage`, which
// renders the accounting and names a suggested id for each check still
// missing. Completeness is thereby a test rather than a claim, and a gap is
// tracked rather than assumed away.
//
// Analyze takes the rules to measure as an argument rather than calling
// checker.GetAllRules itself, so this package needs only the rule model and
// the enumeration, never the code that runs the checks.
package coverage
