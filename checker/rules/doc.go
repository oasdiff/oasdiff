// Package rules defines the metadata model of the backward-compatibility
// rules: the severity levels, the Direction/Area/Kind taxonomy, the Effect a
// rule concludes about the accepted-value set, the Guards that select the
// document state it applies to, and the claims tying each rule to the edits
// it covers (see checker/metaschema).
//
// The model is what makes the rules auditable rather than merely runnable:
// checker/coverage measures the claims against every possible edit, and the
// checker's tests derive each rule's severity from its Effect, Guards and
// Direction and look for broken symmetries across the taxonomy.
//
// The package holds no check implementations: the checker package binds each
// Rule to the function that implements it.
package rules
