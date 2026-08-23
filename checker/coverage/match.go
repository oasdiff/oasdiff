package coverage

import "github.com/oasdiff/oasdiff/checker/metaschema"

// firstMatch returns the index of the first pattern matching the edit. The
// pattern tables are read in their listed order, so a specific pattern must
// precede a general one, which would otherwise absorb its edits and leave it
// accounting for none.
func firstMatch(edit metaschema.Edit, n int, pattern func(int) string) (int, bool) {
	for i := range n {
		if matches, _ := metaschema.MatchPattern(pattern(i), edit); matches {
			return i, true
		}
	}
	return 0, false
}

// matchWaiver returns the waiver that accounts for an unchecked edit.
func matchWaiver(edit metaschema.Edit) (Waiver, bool) {
	if i, ok := firstMatch(edit, len(Waivers), func(i int) string { return Waivers[i].Pattern }); ok {
		return Waivers[i], true
	}
	return Waiver{}, false
}
