// Package rules defines the metadata model of the backward-compatibility
// rules: the severity levels, the Direction/Area/Kind taxonomy, the Effect a
// rule concludes about the accepted-value set, the Guards that select its
// document-state sub-case, and the location claims tying it to the edit
// space in checker/metaschema.
//
// The package holds no check implementations: the checker package binds each
// Rule to its handler and audits the model (coverage, symmetry, and the
// severity law) in its tests.
package rules
