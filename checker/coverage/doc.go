// Package coverage compares the checker's rules against every possible edit
// of an OpenAPI document (see checker/metaschema) and reports what the audit
// decides about each edit: which checks cover it, or why none are expected.
// It is the analysis behind `oasdiff checks changelog coverage` and behind
// the build-failing completeness guard in the checker's tests.
//
// The analysis takes the rules as input rather than reading the registry, so
// it depends on the rule model alone and not on the checker engine.
package coverage
