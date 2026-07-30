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
// the major version increase semver requires, whether the version stayed put,
// moved backwards, or moved by too little. It runs after the checks because
// its input is their output, and before the level filter so it sees changes of
// every level.
//
// It speaks only when a breaking change is present and both versions parse as
// semver: nothing else defines what a major bump is, so a spec versioned by
// date or by a bare "v1" is left alone rather than guessed at. Teams that want
// the policy enforced raise these ids with --severity-levels and gate with
// --fail-on, like any other rule.
func applyVersioningPolicy(config *Config, diffReport *diff.Diff, result Changes) Changes {
	if diffReport.Versions == nil {
		return result
	}

	base, baseOk := parseSemver(diffReport.Versions.Base)
	revision, revisionOk := parseSemver(diffReport.Versions.Revision)
	if !baseOk || !revisionOk {
		return result
	}

	if !hasBreakingChange(config, result) {
		return result
	}

	id, args := versioningViolation(diffReport.Versions, base, revision)
	if id == "" {
		return result
	}

	return append(result, InfoChange{
		Id:    id,
		Level: config.getLogLevel(id),
		Args:  args,
	})
}

// versioningViolation names how this version pair fails to carry a breaking
// change, or returns an empty id when it carries one correctly.
func versioningViolation(versions *diff.VersionPair, base, revision semver) (string, []any) {
	switch {
	case versions.Base == versions.Revision:
		return APIVersionNotBumpedId, []any{versions.Revision}
	case !increased(base, revision):
		return APIVersionDecreasedId, []any{versions.Base, versions.Revision}
	case !majorBumped(base, revision):
		return APIMajorVersionNotBumpedId, []any{versions.Base, versions.Revision}
	}
	return "", nil
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
// Claimed changes are already gone by the time this runs (see the caller).
func hasBreakingChange(config *Config, result Changes) bool {
	for _, change := range result {
		if config.getLogLevel(change.GetId()) >= ERR {
			return true
		}
	}
	return false
}
