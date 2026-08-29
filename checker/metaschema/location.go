package metaschema

import "strings"

// MatchLocation reports whether a rule's location pattern matches an
// edit's location. Pattern segments: "**" matches any run of segments
// (including none), "*" matches exactly one segment, anything else matches
// literally (so "*" in an edit's location is matched by "*" or "**" in the
// pattern).
func MatchLocation(pattern, location string) bool {
	return matchSegments(strings.Split(pattern, "."), strings.Split(location, "."))
}

func matchSegments(pat, loc []string) bool {
	if len(pat) == 0 {
		return len(loc) == 0
	}
	if pat[0] == "**" {
		if matchSegments(pat[1:], loc) {
			return true
		}
		if len(loc) == 0 {
			return false
		}
		return matchSegments(pat, loc[1:])
	}
	if len(loc) == 0 {
		return false
	}
	if pat[0] != "*" && pat[0] != loc[0] {
		return false
	}
	return matchSegments(pat[1:], loc[1:])
}

// MatchPattern reports whether a waiver-style pattern (a location glob with
// an optional ":action[,action...]" restriction) covers the edit.
func MatchPattern(pattern string, edit Edit) (bool, error) {
	if !strings.Contains(pattern, ":") {
		return MatchLocation(pattern, edit.Location), nil
	}
	claim, err := ParseClaim(pattern)
	if err != nil {
		return false, err
	}
	return claim.Matches(edit), nil
}

// joinLocation appends a segment to a location path, the constructing
// counterpart of the splitting MatchLocation does.
func joinLocation(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}
