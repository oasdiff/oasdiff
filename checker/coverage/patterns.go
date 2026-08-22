package coverage

import "github.com/oasdiff/oasdiff/checker/metaschema"

// Pattern is one waiver or non-contract entry with the number of
// edits it accounts for.
type Pattern struct {
	// Kind is "waiver" (what the checks miss today) or "non-contract" (a
	// fact about the object model).
	Kind string `json:"kind" yaml:"kind"`
	// Category refines a waiver: open, resolved-at-usage, or covered-as.
	Category WaiverCategory `json:"category,omitempty" yaml:"category,omitempty"`
	Pattern  string         `json:"pattern" yaml:"pattern"`
	// Edits is the number of edits the entry accounts for; attribution is
	// first-match, in table order.
	Edits  int    `json:"edits" yaml:"edits"`
	Reason string `json:"reason" yaml:"reason"`
}

// Patterns lists the waiver and non-contract entries with the number of
// edits each accounts for.
func Patterns() []Pattern {
	waiverCounts := make([]int, len(Waivers))
	nonContractCounts := make([]int, len(metaschema.NonContracts))
	for _, edit := range metaschema.Edits() {
		if edit.Annotation || edit.Extension {
			continue
		}
		if i, ok := firstMatch(edit, len(metaschema.NonContracts), func(i int) string { return metaschema.NonContracts[i].Pattern }); ok {
			nonContractCounts[i]++
			continue
		}
		if i, ok := firstMatch(edit, len(Waivers), func(i int) string { return Waivers[i].Pattern }); ok {
			waiverCounts[i]++
		}
	}

	result := make([]Pattern, 0, len(Waivers)+len(metaschema.NonContracts))
	for i, w := range Waivers {
		result = append(result, Pattern{Kind: "waiver", Category: w.Category, Pattern: w.Pattern, Edits: waiverCounts[i], Reason: w.Reason})
	}
	for i, nc := range metaschema.NonContracts {
		result = append(result, Pattern{Kind: "non-contract", Pattern: nc.Pattern, Edits: nonContractCounts[i], Reason: nc.Reason})
	}
	return result
}
