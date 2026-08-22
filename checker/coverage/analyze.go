package coverage

import (
	"sort"

	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/checker/rules"
)

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
