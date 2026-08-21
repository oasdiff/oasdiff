package checker

import (
	"sort"
	"strings"

	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// CoverageStatus is what the audit decided about one possible edit.
type CoverageStatus string

const (
	// CoverageCovered: at least one check claims the edit.
	CoverageCovered CoverageStatus = "covered"
	// CoverageUncovered: a wire-relevant edit with no check and no waiver;
	// the audit fails the build until it gains one or the other.
	CoverageUncovered CoverageStatus = "uncovered"
	// CoverageWaived: a wire-relevant edit with no check, accounted for by a
	// coverage waiver.
	CoverageWaived CoverageStatus = "waived"
	// CoverageNonContract: the edit cannot change which payloads are valid
	// (an annotation, a specification extension, or a metaschema.NonContracts
	// entry), so no check is expected.
	CoverageNonContract CoverageStatus = "non-contract"
)

// CoverageEdit is one possible edit of an OpenAPI document with what the
// audit decided about it: its status, the checks that cover it, or the
// reason none are expected.
type CoverageEdit struct {
	Location string         `json:"location" yaml:"location"`
	Action   string         `json:"action" yaml:"action"`
	Polarity string         `json:"polarity" yaml:"polarity"`
	Status   CoverageStatus `json:"status" yaml:"status"`
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

// Coverage maps every possible edit of an OpenAPI document to what the
// audit decided about it, sorted by location then action.
func Coverage() []CoverageEdit {
	type claimant struct {
		claim metaschema.Claim
		id    string
	}
	var claims []claimant
	for _, rule := range GetAllRules() {
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				continue // the checker's TestRuleLocations reports it
			}
			claims = append(claims, claimant{claim, rule.Id})
		}
	}

	var result []CoverageEdit
	for _, edit := range metaschema.Edits() {
		row := CoverageEdit{
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
			row.Status = CoverageCovered
			sort.Strings(row.Checks)
		case edit.Annotation:
			row.Status = CoverageNonContract
			row.Reason = "annotation: documentation-only field"
		case edit.Extension:
			row.Status = CoverageNonContract
			row.Reason = "specification extension"
		default:
			if reason, ok := nonContractReason(edit); ok {
				row.Status = CoverageNonContract
				row.Reason = reason
			} else if waiver, ok := matchWaiver(edit); ok {
				row.Status = CoverageWaived
				row.Category = waiver.Category
				row.Reason = waiver.Reason
				// only an open waiver implies a missing check; the other
				// categories say the edit is handled elsewhere
				if waiver.Category == CategoryOpen {
					row.SuggestedId = suggestId(edit)
				}
			} else {
				row.Status = CoverageUncovered
				row.SuggestedId = suggestId(edit)
			}
		}
		result = append(result, row)
	}
	return result
}

// CoveragePattern is one waiver or non-contract entry with the number of
// edits it accounts for.
type CoveragePattern struct {
	// Kind is "waiver" (relative to the rule registry) or "non-contract"
	// (a fact about the object model).
	Kind string `json:"kind" yaml:"kind"`
	// Category refines a waiver: open, resolved-at-usage, or covered-as.
	Category WaiverCategory `json:"category,omitempty" yaml:"category,omitempty"`
	Pattern  string         `json:"pattern" yaml:"pattern"`
	// Edits is the number of edits the entry accounts for; attribution is
	// first-match, in table order.
	Edits  int    `json:"edits" yaml:"edits"`
	Reason string `json:"reason" yaml:"reason"`
}

// CoveragePatterns lists the waiver and non-contract entries with the number
// of edits each accounts for.
func CoveragePatterns() []CoveragePattern {
	waiverCounts := make([]int, len(CoverageWaivers))
	nonContractCounts := make([]int, len(metaschema.NonContracts))
	for _, edit := range metaschema.Edits() {
		if edit.Annotation || edit.Extension {
			continue
		}
		if i, ok := firstMatch(edit, len(metaschema.NonContracts), func(i int) string { return metaschema.NonContracts[i].Pattern }); ok {
			nonContractCounts[i]++
			continue
		}
		if i, ok := firstMatch(edit, len(CoverageWaivers), func(i int) string { return CoverageWaivers[i].Pattern }); ok {
			waiverCounts[i]++
		}
	}

	result := make([]CoveragePattern, 0, len(CoverageWaivers)+len(metaschema.NonContracts))
	for i, w := range CoverageWaivers {
		result = append(result, CoveragePattern{Kind: "waiver", Category: w.Category, Pattern: w.Pattern, Edits: waiverCounts[i], Reason: w.Reason})
	}
	for i, nc := range metaschema.NonContracts {
		result = append(result, CoveragePattern{Kind: "non-contract", Pattern: nc.Pattern, Edits: nonContractCounts[i], Reason: nc.Reason})
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

func matchWaiver(edit metaschema.Edit) (CoverageWaiver, bool) {
	if i, ok := firstMatch(edit, len(CoverageWaivers), func(i int) string { return CoverageWaivers[i].Pattern }); ok {
		return CoverageWaivers[i], true
	}
	return CoverageWaiver{}, false
}

// suggestId derives a candidate check id for an unchecked edit, in the
// naming style of the existing ids: a context prefix from the location, the
// edited keyword, and the action as a past-tense verb. It is a hint for
// naming the missing check, in the spirit of the retired generator (#1168).
func suggestId(edit metaschema.Edit) string {
	verb := map[metaschema.Action]string{
		metaschema.ActionAdd:      "added",
		metaschema.ActionRemove:   "removed",
		metaschema.ActionChange:   "changed",
		metaschema.ActionSet:      "set",
		metaschema.ActionUnset:    "unset",
		metaschema.ActionIncrease: "increased",
		metaschema.ActionDecrease: "decreased",
	}[edit.Action]

	loc := edit.Location
	var prefix string
	switch {
	case strings.HasPrefix(loc, "components.securitySchemes"):
		prefix = "api-security-scheme"
	case strings.HasPrefix(loc, "components"):
		prefix = "api-component"
	case strings.HasPrefix(loc, "webhooks"):
		prefix = "webhook"
	case strings.HasPrefix(loc, "security"):
		prefix = "api-security"
	case strings.Contains(loc, ".callbacks."):
		prefix = "callback"
	case strings.Contains(loc, ".responses.") && strings.Contains(loc, ".headers."):
		prefix = "response-header"
	case strings.Contains(loc, ".responses."):
		prefix = "response"
	case strings.Contains(loc, ".requestBody"):
		prefix = "request-body"
	case strings.Contains(loc, ".parameters."):
		prefix = "request-parameter"
	case strings.HasPrefix(loc, "paths."):
		prefix = "api"
	default:
		prefix = strings.SplitN(loc, ".", 2)[0]
	}

	// the edited keyword: the last segment that names a field; a map entry
	// is named by its parent, singular
	segs := strings.Split(loc, ".")
	keyword := segs[len(segs)-1]
	if keyword == "*" && len(segs) > 1 {
		keyword = strings.TrimSuffix(segs[len(segs)-2], "s")
	}
	keyword = strings.TrimPrefix(keyword, "$")

	id := prefix + "-" + toKebab(keyword) + "-" + verb
	return strings.ReplaceAll(id, "body-body", "body")
}

func toKebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
