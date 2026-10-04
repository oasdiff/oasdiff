package consolidate

import "github.com/oasdiff/oasdiff/checker"

// Mode merges the findings it recognizes as reporting one change, and passes
// the rest through.
type Mode func(checker.Changes) checker.Changes

// Changes applies modes in order.
func Changes(changes checker.Changes, modes ...Mode) checker.Changes {
	for _, mode := range modes {
		changes = mode(changes)
	}
	return changes
}
