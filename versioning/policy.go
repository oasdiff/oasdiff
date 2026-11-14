package versioning

import (
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
)

// Reason names how a version pair fails to carry a breaking change. The
// policy is one statement, a breaking change requires a major version
// increase, and these are the three ways to violate it.
type Reason int

const (
	// NoReason: nothing was violated.
	NoReason Reason = iota
	// NotBumped: info.version is unchanged.
	NotBumped
	// Decreased: info.version moved backwards.
	Decreased
	// MajorNotBumped: info.version moved, but the major version did not
	// increase.
	MajorNotBumped
)

// Status is the verdict on one comparison.
type Status int

const (
	// NotApplicable: no breaking change, so no version increase is required.
	NotApplicable Status = iota
	// Undecidable: a breaking change, but the two versions don't define what
	// a major increase is.
	Undecidable
	Satisfied
	Violated
)

// Verdict answers whether info.version carries the breaking changes a
// comparison found. Reason names the violation, and is NoReason for every
// other status.
type Verdict struct {
	Status          Status
	Reason          Reason
	BaseVersion     string
	RevisionVersion string
	BreakingChanges int
}

// Check judges info.version against the breaking changes in changes: whether
// the version stayed put, moved backwards, moved by too little, or carried
// them correctly. The caller decides what the verdict is worth, which for the
// CLI is --fail-on MISSING-VERSION-BUMP.
func Check(diffReport *diff.Diff, changes checker.Changes) Verdict {
	breaking := countBreakingChanges(changes)
	if breaking == 0 {
		return Verdict{Status: NotApplicable}
	}

	verdict := Verdict{
		Status:          Undecidable,
		BreakingChanges: breaking,
	}

	if diffReport == nil || diffReport.BaseInfo == nil || diffReport.RevisionInfo == nil {
		return verdict
	}
	verdict.BaseVersion = diffReport.BaseInfo.Version
	verdict.RevisionVersion = diffReport.RevisionInfo.Version

	base, baseOk := parseSemver(verdict.BaseVersion)
	revision, revisionOk := parseSemver(verdict.RevisionVersion)
	if !baseOk || !revisionOk {
		return verdict
	}

	if reason := versioningViolation(verdict.BaseVersion, verdict.RevisionVersion, base, revision); reason != NoReason {
		verdict.Status = Violated
		verdict.Reason = reason
		return verdict
	}

	verdict.Status = Satisfied
	return verdict
}

// versioningViolation names how this version pair fails to carry a breaking
// change, or returns NoReason when it carries one correctly.
func versioningViolation(baseVersion, revisionVersion string, base, revision semver) Reason {
	switch {
	case baseVersion == revisionVersion:
		return NotBumped
	case !increased(base, revision):
		return Decreased
	case !majorBumped(base, revision):
		return MajorNotBumped
	}
	return NoReason
}

// majorBumped reports whether the version moved by enough to carry a breaking
// change. Below 1.0.0 semver gives the minor the major's role ("anything may
// change at any time"), so a minor bump satisfies the policy there; demanding
// that every breaking change take a 0.x spec to 1.0.0 would flag most
// pre-release specs on every change.
func majorBumped(base, revision semver) bool {
	if base.major == 0 && revision.major == 0 {
		return revision.minor > base.minor
	}
	return revision.major > base.major
}

// A change counts at the level it is reported at, which an override, a
// disclaimer or an ignore file may have moved.
func countBreakingChanges(changes checker.Changes) int {
	count := 0
	for _, change := range changes {
		if change.GetLevel() >= checker.ERR {
			count++
		}
	}
	return count
}
