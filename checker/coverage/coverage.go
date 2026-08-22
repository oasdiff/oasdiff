package coverage

import (
	"sort"

	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/checker/rules"
)

// Status is what the audit decided about one possible edit.
type Status string

const (
	// Covered: at least one check claims the edit.
	Covered Status = "covered"
	// Uncovered: a wire-relevant edit with no check and no waiver;
	// the audit fails the build until it gains one or the other.
	Uncovered Status = "uncovered"
	// Waived: a wire-relevant edit with no check, accounted for by a
	// coverage waiver.
	Waived Status = "waived"
	// NonContract: the edit cannot change which payloads are valid
	// (an annotation, a specification extension, or a metaschema.NonContracts
	// entry), so no check is expected.
	NonContract Status = "non-contract"
)

// Edit is one possible edit of an OpenAPI document with what the
// audit decided about it: its status, the checks that cover it, or the
// reason none are expected.
type Edit struct {
	Location string `json:"location" yaml:"location"`
	Action   string `json:"action" yaml:"action"`
	Polarity string `json:"polarity" yaml:"polarity"`
	Status   Status `json:"status" yaml:"status"`
	// Category refines a waived status: open (a missing check), or handled
	// elsewhere (resolved-at-usage, covered-as).
	Category WaiverCategory `json:"category,omitempty" yaml:"category,omitempty"`
	// Checks are the ids of the checks claiming the edit (covered only).
	Checks []string `json:"checks,omitempty" yaml:"checks,omitempty"`
	// Reason explains a waived or non-contract status.
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
	// SuggestedId is a derived candidate check id for an uncovered or waived
	// edit: a naming hint for the missing check, not a promise of one.
	SuggestedId string `json:"suggestedId,omitempty" yaml:"suggestedId,omitempty"`
}

// Analyze maps every possible edit of an OpenAPI document to what the
// audit decides about it given these rules, sorted by location then action.
func Analyze(ruleset []rules.Rule) []Edit {
	type claimant struct {
		claim metaschema.Claim
		id    string
	}
	var claims []claimant
	for _, rule := range ruleset {
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				continue // an invalid claim fails the checker's audit; not reported here
			}
			claims = append(claims, claimant{claim, rule.Id})
		}
	}

	var result []Edit
	for _, edit := range metaschema.Edits() {
		row := Edit{
			Location: edit.Location,
			Action:   string(edit.Action),
			Polarity: string(edit.Polarity),
		}
		for _, c := range claims {
			if c.claim.Matches(edit) {
				row.Checks = append(row.Checks, c.id)
			}
		}
		switch {
		case len(row.Checks) > 0:
			row.Status = Covered
			sort.Strings(row.Checks)
		case edit.Annotation:
			row.Status = NonContract
			row.Reason = "annotation: documentation-only field"
		case edit.Extension:
			row.Status = NonContract
			row.Reason = "specification extension"
		default:
			if reason, ok := nonContractReason(edit); ok {
				row.Status = NonContract
				row.Reason = reason
			} else if waiver, ok := matchWaiver(edit); ok {
				row.Status = Waived
				row.Category = waiver.Category
				row.Reason = waiver.Reason
				// only an open waiver implies a missing check; the other
				// categories say the edit is handled elsewhere
				if waiver.Category == CategoryOpen {
					row.SuggestedId = suggestId(edit)
				}
			} else {
				row.Status = Uncovered
				row.SuggestedId = suggestId(edit)
			}
		}
		result = append(result, row)
	}
	return result
}

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
