// Package versioning judges a comparison against the versioning policy: a
// breaking change requires a major version increase in info.version.
//
// Check takes the changes a checker run produced and the diff they came from,
// and returns a Verdict: whether the policy applies, whether the two versions
// can be compared at all, and if they can, whether the version movement
// carries the breaking changes.
package versioning
