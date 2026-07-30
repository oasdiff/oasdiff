package checker

import (
	"regexp"
	"strconv"

	"github.com/oasdiff/oasdiff/diff"
)

const APIMajorVersionNotBumpedId = "api-major-version-not-bumped"

// semverPattern is the official semver.org grammar, anchored, with an
// optional leading "v". Versions that don't match are not judged at all (see
// applyVersioningPolicy), so the pattern must not accept anything looser: a
// date, or a bare "v1" read as 1.0.0, would produce confident findings about a
// versioning scheme oasdiff was never told to enforce.
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

// applyVersioningPolicy reports a breaking change that was released without
// the major-version bump semver requires. It runs after the checks because its
// input is their output, and before the level filter so it sees changes of
// every level.
//
// Three conditions must all hold, which is what makes it quiet enough to need
// no opt-in flag:
//
//   - info.version changed. An unchanged version is not evidence of a missed
//     bump: it is indistinguishable from a spec that does not use info.version
//     as a release signal at all, which is the common case (of 174 spec pairs
//     in this repo's own test data, three change the version). Reporting it
//     would fire on nearly every user on nearly every change.
//   - Both versions parse as semver. Nothing else defines what a major bump is.
//   - A breaking change was found, at the levels the user configured.
//
// Teams that want this enforced raise the id with --severity-levels and gate
// with --fail-on, like any other rule.
func applyVersioningPolicy(config *Config, diffReport *diff.Diff, result Changes) Changes {
	if diffReport.Versions == nil || diffReport.Versions.Base == diffReport.Versions.Revision {
		return result
	}

	base, baseOk := parseSemver(diffReport.Versions.Base)
	revision, revisionOk := parseSemver(diffReport.Versions.Revision)
	if !baseOk || !revisionOk {
		return result
	}

	// A version that moved backwards is a different question, and one this
	// rule does not answer (see the symmetry waiver for api-major-version-not-bumped):
	// judging it as an insufficient bump would be wrong, since nothing was bumped.
	if !increased(base, revision) {
		return result
	}

	if majorBumped(base, revision) || !hasBreakingChange(config, result) {
		return result
	}

	return append(result, InfoChange{
		Id:    APIMajorVersionNotBumpedId,
		Level: config.getLogLevel(APIMajorVersionNotBumpedId),
		Args:  []any{diffReport.Versions.Base, diffReport.Versions.Revision},
	})
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

// hasBreakingChange reports whether any change reaches ERR at the configured
// levels, so a team that has downgraded a check has downgraded it here too.
// Claimed changes are excluded because the runner drops them, and this rule is
// excluded so it never judges itself.
func hasBreakingChange(config *Config, result Changes) bool {
	for _, change := range result {
		if apiChange, ok := change.(ApiChange); ok && apiChange.claimed {
			continue
		}
		if change.GetId() == APIMajorVersionNotBumpedId {
			continue
		}
		if config.getLogLevel(change.GetId()) >= ERR {
			return true
		}
	}
	return false
}
