package internal

import (
	"fmt"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/versioning"
)

// versioningPolicyLine states the verdict and the reason for it. A pass needs
// the reason as much as a failure does: an exit code of 0 with breaking
// changes on screen reads as a bug without one.
func versioningPolicyLine(verdict versioning.Verdict) string {
	switch verdict.Status {
	case versioning.Satisfied:
		return fmt.Sprintf("versioning policy: satisfied. info.version moved from %s to %s, which allows the breaking changes.",
			verdict.BaseVersion, verdict.RevisionVersion)
	case versioning.Violated:
		return fmt.Sprintf("versioning policy: violated. %s.", versionMovement(verdict))
	case versioning.Undecidable:
		return fmt.Sprintf("versioning policy: cannot be evaluated. info.version must be a semantic version in both specs (base %q, revision %q).",
			verdict.BaseVersion, verdict.RevisionVersion)
	}
	return "versioning policy: not applicable. No breaking changes, so no version increase is required."
}

// versionMovement describes what info.version did instead of carrying the
// breaking changes.
func versionMovement(verdict versioning.Verdict) string {
	switch verdict.Reason {
	case versioning.NotBumped:
		return fmt.Sprintf("info.version is still %s", verdict.RevisionVersion)
	case versioning.Decreased:
		return fmt.Sprintf("info.version decreased from %s to %s", verdict.BaseVersion, verdict.RevisionVersion)
	}
	return fmt.Sprintf("info.version moved from %s to %s without a major version increase",
		verdict.BaseVersion, verdict.RevisionVersion)
}

// failOnVersionBump decides the exit code under --enforce-version-bump: the
// policy answers for errors, and failOn, the --fail-on level, still answers
// for findings below error, which a major version increase says nothing
// about. NONE means the caller asked about nothing else.
func failOnVersionBump(failOn checker.Level, errs checker.Changes, verdict versioning.Verdict) bool {
	switch verdict.Status {
	case versioning.Violated:
		return true
	case versioning.Undecidable:
		// Enforcement was asked for and the versions don't allow a verdict, so
		// fail rather than pass a release the policy never judged.
		return true
	}

	if failOn == checker.NONE {
		return false
	}

	if verdict.Status == versioning.NotApplicable {
		return errs.HasLevelOrHigher(failOn)
	}

	return hasLevelBelowErr(errs, failOn)
}

func hasLevelBelowErr(errs checker.Changes, level checker.Level) bool {
	for _, change := range errs {
		if changeLevel := change.GetLevel(); changeLevel >= level && changeLevel < checker.ERR {
			return true
		}
	}
	return false
}
