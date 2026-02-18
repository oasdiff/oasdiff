package checker

import "cmp"

type Changes []Change

func (changes Changes) HasLevelOrHigher(level Level) bool {
	for _, change := range changes {
		if change.GetLevel() >= level {
			return true
		}
	}
	return false
}

func (changes Changes) GetLevelCount() map[Level]int {
	counts := map[Level]int{}
	for _, change := range changes {
		level := change.GetLevel()
		counts[level] = counts[level] + 1
	}
	return counts
}

func CompareChanges(a, b Change) int {
	// Level descending (most severe first)
	if c := cmp.Compare(a.GetLevel(), b.GetLevel()); c != 0 {
		return -c
	}
	if c := cmp.Compare(a.GetPath(), b.GetPath()); c != 0 {
		return c
	}
	if c := cmp.Compare(a.GetOperation(), b.GetOperation()); c != 0 {
		return c
	}
	if c := cmp.Compare(a.GetId(), b.GetId()); c != 0 {
		return c
	}
	if c := cmp.Compare(len(a.GetArgs()), len(b.GetArgs())); c != 0 {
		return c
	}
	for i, arg := range a.GetArgs() {
		ia := interfaceToString(arg)
		ja := interfaceToString(b.GetArgs()[i])
		if c := cmp.Compare(ia, ja); c != 0 {
			return c
		}
	}
	return 0
}
