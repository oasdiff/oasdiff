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
		return fmt.Sprintf("versioning policy: satisfied. info.version moved from %s to %s, which allows the %s.",
			verdict.BaseVersion, verdict.RevisionVersion, breakingChanges(verdict.BreakingChanges))
	case versioning.Violated:
		return fmt.Sprintf("versioning policy: violated (%s). %s, and %s.",
			verdict.Reason, breakingChanges(verdict.BreakingChanges), versionMovement(verdict))
	case versioning.Undecidable:
		return fmt.Sprintf("versioning policy: cannot be evaluated. info.version must be a semantic version in both specs (base %q, revision %q).",
			verdict.BaseVersion, verdict.RevisionVersion)
	}
	return "versioning policy: not applicable. No breaking changes, so no version increase is required."
}

func breakingChanges(count int) string {
	if count == 1 {
		return "1 breaking change"
	}
	return fmt.Sprintf("%d breaking changes", count)
}

// versionMovement describes what info.version did instead of carrying the
// breaking changes.
func versionMovement(verdict versioning.Verdict) string {
	switch verdict.Reason {
	case versioning.ReasonNotBumped:
		return fmt.Sprintf("info.version is still %s", verdict.RevisionVersion)
	case versioning.ReasonDecreased:
		return fmt.Sprintf("info.version decreased from %s to %s", verdict.BaseVersion, verdict.RevisionVersion)
	}
	return fmt.Sprintf("info.version moved from %s to %s without a major version increase",
		verdict.BaseVersion, verdict.RevisionVersion)
}

// failOnVersionBump decides the exit code under --enforce-version-bump: the
// policy answers for errors, and --fail-on still answers for findings below
// error, which a major version increase does not license.
func failOnVersionBump(flags *Flags, errs checker.Changes, verdict versioning.Verdict) (bool, *ReturnError) {
	switch verdict.Status {
	case versioning.Violated:
		return true, nil
	case versioning.Undecidable:
		// Enforcement was asked for and the versions don't allow a verdict, so
		// fail rather than pass a release the policy never judged.
		return true, nil
	}

	if flags.getFailOn() == "" {
		return false, nil
	}

	level, returnErr := flags.failOnLevel()
	if returnErr != nil {
		return false, returnErr
	}

	if verdict.Status == versioning.NotApplicable {
		return errs.HasLevelOrHigher(level), nil
	}

	return hasLevelBelowErr(errs, level), nil
}

func hasLevelBelowErr(errs checker.Changes, level checker.Level) bool {
	for _, change := range errs {
		if changeLevel := change.GetLevel(); changeLevel >= level && changeLevel < checker.ERR {
			return true
		}
	}
	return false
}
