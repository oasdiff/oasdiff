package coverage

import (
	"sort"

	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/checker/rules"
)

// Analyze maps every possible edit of an OpenAPI document to what the
// audit decides about it given these rules, sorted by location then action.
func Analyze(ruleset []rules.Rule) []Edit {
	claims := parseClaims(ruleset)

	edits := metaschema.Edits()
	result := make([]Edit, len(edits))
	for i, edit := range edits {
		result[i] = decide(edit, claims)
	}
	return result
}

// claimant is one parsed claim with the check that makes it.
type claimant struct {
	claim metaschema.Claim
	id    string
}

func parseClaims(ruleset []rules.Rule) []claimant {
	var claims []claimant
	for _, rule := range ruleset {
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				// a claim that does not parse covers nothing, so dropping
				// it can only understate coverage, never overstate it
				continue
			}
			claims = append(claims, claimant{claim, rule.Id})
		}
	}
	return claims
}

// decide reports what the audit makes of one edit: the checks that claim
// it, or the reason none are expected.
func decide(edit metaschema.Edit, claims []claimant) Edit {
	row := Edit{
		Location: edit.Location,
		Action:   string(edit.Action),
		Polarity: string(edit.Polarity),
		Checks:   claimingChecks(edit, claims),
	}

	if len(row.Checks) > 0 {
		row.Status = Covered
		return row
	}

	if reason, ok := metaschema.NonContractReason(edit); ok {
		row.Status = NonContract
		row.Reason = reason
		return row
	}

	if waiver, ok := matchWaiver(edit); ok {
		row.Status = Waived
		row.Category = waiver.Category
		row.Reason = waiver.Reason
		// only an open waiver implies a missing check; the other categories
		// say the edit is handled elsewhere
		if waiver.Category == CategoryOpen {
			row.SuggestedId = suggestId(edit)
		}
		return row
	}

	row.Status = Uncovered
	row.SuggestedId = suggestId(edit)
	return row
}

// claimingChecks returns the ids of the checks whose claims cover the edit.
func claimingChecks(edit metaschema.Edit, claims []claimant) []string {
	var ids []string
	for _, c := range claims {
		if c.claim.Matches(edit) {
			ids = append(ids, c.id)
		}
	}
	sort.Strings(ids)
	return ids
}
