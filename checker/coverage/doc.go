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
// Package coverage answers the question. It measures the rules against every
// possible edit of an OpenAPI document (see checker/metaschema) and reports
// what the audit decides about each edit: which checks cover it, or why none
// are expected. Each rule declares the edits it covers, and every edit that
// can change which payloads are valid must be covered by a check or waived
// with a reviewed reason. Completeness thereby stops being a claim and
// becomes a test: an uncovered edit fails the build, and so does a waiver
// that no longer accounts for anything, so a gap is identified and tracked
// rather than assumed away. `oasdiff checks changelog coverage` renders the
// accounting, and lists a suggested id for each check still missing.
//
// Analyze takes the rules to measure as an argument rather than reading the
// checker's registry, so this package needs only the rule model and the
// enumeration, never the code that runs the checks.
package coverage
