package checker

import (
	"regexp"
	"strconv"

	"github.com/oasdiff/oasdiff/diff"
)

// The versioning policy is one statement: a breaking change requires a major
// version increase. These are the three ways to violate it.
const (
	APIVersionNotBumpedId      = "api-version-not-bumped"
	APIVersionDecreasedId      = "api-version-decreased"
	APIMajorVersionNotBumpedId = "api-major-version-not-bumped"
)

// VersioningPolicyStatus is the verdict on one comparison.
type VersioningPolicyStatus int

const (
	// VersioningPolicyNotApplicable: no breaking change, so no version
	// increase is required.
	VersioningPolicyNotApplicable VersioningPolicyStatus = iota
	// VersioningPolicyUndecidable: a breaking change, but the two versions
	// don't define what a major increase is.
	VersioningPolicyUndecidable
	VersioningPolicySatisfied
	VersioningPolicyViolated
)

// VersioningVerdict answers whether info.version carries the breaking changes
// a comparison found. Reason names the violation, and is empty for every other
// status.
type VersioningVerdict struct {
	Status          VersioningPolicyStatus
	Reason          string
	BaseVersion     string
	RevisionVersion string
	BreakingChanges int
}

// semverPattern is the official semver.org grammar, anchored, with an
// optional leading "v". Versions that don't match leave the policy with no
// definition of a major increase (see CheckVersioningPolicy), so the pattern
// must not accept anything looser: a date, or a bare "v1" read as 1.0.0, would
// produce a confident verdict about a versioning scheme oasdiff was never told
// to enforce.
var semverPattern = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

type semver struct {
	major, minor, patch int
}

// parseSemver reports whether version is a semantic version, and its numeric
// precedence fields. Prerelease and build metadata are accepted and ignored:
// the policy only asks which number moved.
func parseSemver(version string) (semver, bool) {
	m := semverPattern.FindStringSubmatch(version)
	if m == nil {
		return semver{}, false
	}
	// Each group matched \d+ under the anchored pattern, so Atoi fails only on
	// integer overflow, which is not a version anyone enforces policy on.
	major, err1 := strconv.Atoi(m[1])
	minor, err2 := strconv.Atoi(m[2])
	patch, err3 := strconv.Atoi(m[3])
	if err1 != nil || err2 != nil || err3 != nil {
		return semver{}, false
	}
	return semver{major, minor, patch}, true
}

// CheckVersioningPolicy judges info.version against the breaking changes in
// changes: whether the version stayed put, moved backwards, moved by too
// little, or carried them correctly.
//
// It is not a check, and adds nothing to changes. The caller decides what a
// verdict is worth, which for the CLI is --enforce-version-bump.
func CheckVersioningPolicy(diffReport *diff.Diff, changes Changes) VersioningVerdict {
	breaking := countBreakingChanges(changes)
	if breaking == 0 {
		return VersioningVerdict{Status: VersioningPolicyNotApplicable}
	}

	verdict := VersioningVerdict{
		Status:          VersioningPolicyUndecidable,
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

	if reason := versioningViolation(verdict.BaseVersion, verdict.RevisionVersion, base, revision); reason != "" {
		verdict.Status = VersioningPolicyViolated
		verdict.Reason = reason
		return verdict
	}

	verdict.Status = VersioningPolicySatisfied
	return verdict
}

// versioningViolation names how this version pair fails to carry a breaking
// change, or returns an empty string when it carries one correctly.
func versioningViolation(baseVersion, revisionVersion string, base, revision semver) string {
	switch {
	case baseVersion == revisionVersion:
		return APIVersionNotBumpedId
	case !increased(base, revision):
		return APIVersionDecreasedId
	case !majorBumped(base, revision):
		return APIMajorVersionNotBumpedId
	}
	return ""
}

// increased reports whether revision comes after base in semver precedence.
// Prerelease and build metadata are not compared: they don't decide which
// number was bumped, which is all this policy asks.
func increased(base, revision semver) bool {
	if revision.major != base.major {
		return revision.major > base.major
	}
	if revision.minor != base.minor {
		return revision.minor > base.minor
	}
	return revision.patch > base.patch
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
func countBreakingChanges(changes Changes) int {
	count := 0
	for _, change := range changes {
		if change.GetLevel() >= ERR {
			count++
		}
	}
	return count
}
