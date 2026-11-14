package versioning

import (
	"regexp"
	"strconv"
)

// semverPattern is the official semver.org grammar, anchored, with an
// optional leading "v". Versions that don't match leave the policy with no
// definition of a major increase (see Check), so the pattern must not accept
// anything looser: a date, or a bare "v1" read as 1.0.0, would produce a
// confident verdict about a versioning scheme oasdiff was never told to
// enforce.
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
