package internal

import (
	"fmt"
	"io"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
)

// versioningPolicyLine states the verdict and the reason for it. A pass needs
// the reason as much as a failure does: an exit code of 0 with breaking
// changes on screen reads as a bug without one.
func versioningPolicyLine(verdict checker.VersioningVerdict) string {
	switch verdict.Status {
	case checker.VersioningPolicySatisfied:
		return fmt.Sprintf("versioning policy: satisfied. info.version moved from %s to %s, which allows the %s.",
			verdict.BaseVersion, verdict.RevisionVersion, breakingChanges(verdict.BreakingChanges))
	case checker.VersioningPolicyViolated:
		return fmt.Sprintf("versioning policy: violated (%s). %s, and %s.",
			verdict.Reason, breakingChanges(verdict.BreakingChanges), versionMovement(verdict))
	case checker.VersioningPolicyUndecidable:
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
func versionMovement(verdict checker.VersioningVerdict) string {
	switch verdict.Reason {
	case checker.APIVersionNotBumpedId:
		return fmt.Sprintf("info.version is still %s", verdict.RevisionVersion)
	case checker.APIVersionDecreasedId:
		return fmt.Sprintf("info.version decreased from %s to %s", verdict.BaseVersion, verdict.RevisionVersion)
	}
	return fmt.Sprintf("info.version moved from %s to %s without a major version increase",
		verdict.BaseVersion, verdict.RevisionVersion)
}

// failOnVersionBump decides the exit code under --enforce-version-bump: the
// policy answers for errors, and --fail-on still answers for findings below
// error, which a major version increase does not license.
func failOnVersionBump(flags *Flags, errs checker.Changes, verdict checker.VersioningVerdict) (bool, *ReturnError) {
	switch verdict.Status {
	case checker.VersioningPolicyViolated:
		return true, nil
	case checker.VersioningPolicyUndecidable:
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

	if verdict.Status == checker.VersioningPolicyNotApplicable {
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

// dropRetiredSeverityLevels removes the ids that are no longer checks from
// severityLevels, and returns them sorted. The checker config accepts a level
// only for an id it knows, so these have to go before it is built.
func dropRetiredSeverityLevels(severityLevels map[string]checker.Level) []string {
	retired := checker.RetiredRuleIdsIn(severityLevels)
	for _, id := range retired {
		delete(severityLevels, id)
	}
	return retired
}

// warnRetiredSeverityLevels says that a level set for a retired id has no
// effect, and names the flag that replaced it.
func warnRetiredSeverityLevels(w io.Writer, retired []string, file string) {
	if len(retired) == 0 {
		return
	}

	subject := retired[0] + " is no longer a check"
	if len(retired) > 1 {
		subject = strings.Join(retired, ", ") + " are no longer checks"
	}

	_, _ = fmt.Fprintf(w, "warning: %s, so the level set in %s has no effect. Use --enforce-version-bump to require a major version increase for breaking changes.\n", subject, file)
}
