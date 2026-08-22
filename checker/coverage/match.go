package coverage

import "github.com/oasdiff/oasdiff/checker/metaschema"

// The lookups that decide which entry explains an unchecked edit, shared
// by Analyze and Patterns. Both tables are first-match, in their listed
// order, so a specific pattern must precede a general one.

func firstMatch(edit metaschema.Edit, n int, pattern func(int) string) (int, bool) {
	for i := range n {
		if matches, _ := metaschema.MatchPattern(pattern(i), edit); matches {
			return i, true
		}
	}
	return 0, false
}

func nonContractReason(edit metaschema.Edit) (string, bool) {
	if i, ok := firstMatch(edit, len(metaschema.NonContracts), func(i int) string { return metaschema.NonContracts[i].Pattern }); ok {
		return metaschema.NonContracts[i].Reason, true
	}
	return "", false
}

func matchWaiver(edit metaschema.Edit) (Waiver, bool) {
	if i, ok := firstMatch(edit, len(Waivers), func(i int) string { return Waivers[i].Pattern }); ok {
		return Waivers[i], true
	}
	return Waiver{}, false
}
