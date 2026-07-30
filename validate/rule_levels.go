package validate

import "github.com/oasdiff/oasdiff/checker"

// Severities for the rules `oasdiff validate` can report.
//
// A rule's severity is a property of the rule, not of the individual finding:
// each kin error type carries its own code, so the code determines the
// severity. That makes this the single source of truth for both ends, the
// runtime classification of a finding (severityForKinError) and the static
// listing (`oasdiff checks validate`), which cannot then disagree.
//
// ERR is the default and is not listed: every kin Validate result is a spec
// violation, so an unrecognised or newly added rule stays an error until
// someone decides otherwise. Only the downgrades are recorded here.
//
// Deliberately left at ERR despite being downgrade candidates:
// duplicate-operation-id violates the spec's uniqueness MUST and breaks code
// generators that key method names off operationId.
var ruleLevels = map[string]checker.Level{
	// INFO: an example that doesn't match its schema is a documentation
	// accuracy nit; the contract (the schema) is still valid.
	"example-violates-schema": checker.INFO,

	// WARN: structurally valid, but a portability or correctness risk.
	//
	// A default, unlike an example, is consumed at runtime by some tooling, so
	// a mismatch there is a real risk.
	"default-violates-schema": checker.WARN,
	// Fields alongside a $ref are silently ignored in 3.0 rather than breaking
	// the spec; the author's intent is lost, not the document.
	"extra-sibling-fields": checker.WARN,
	"conflicting-paths":    checker.WARN,
	"duplicate-parameter":  checker.WARN,

	// oasdiff's own SHOULD-level lints, which kin does not enforce. The spec
	// still parses in every case; each is a correctness or portability risk.
	DuplicateEnumValueID:              checker.WARN,
	AmbiguousParameterSerializationID: checker.WARN,
	RequiredWithDefaultID:             checker.WARN,
	TypeFormatMismatchID:              checker.WARN,
}

// RuleLevel returns the severity of a validate rule.
//
// The version-gate rules (`<field>-field-for-3-1-plus` and friends) are WARN
// as a family rather than one entry each, matched the same way
// versionGateDescription describes them: a 3.1-only field in an older document
// is a portability problem, not a structural break. Deriving them keeps a
// newly gated field added upstream classified without a change here.
//
// An unknown id gets ERR, the same default an unrecognised error takes at
// runtime.
func RuleLevel(id string) checker.Level {
	if level, ok := ruleLevels[id]; ok {
		return level
	}
	if versionGateRe.MatchString(id) {
		return checker.WARN
	}
	return checker.ERR
}
