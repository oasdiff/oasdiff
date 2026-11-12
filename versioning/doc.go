// Package versioning judges a comparison against the versioning policy: a
// breaking change requires a major version increase in info.version.
//
// The answer is a Verdict rather than a change, because it is a statement
// about the release and not about the API: it names no endpoint, and a
// caller that wants it asks for it. Check takes the changes a checker run
// produced and the diff they came from, and says whether the version
// movement carries them.
package versioning
