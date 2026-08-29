package coverage

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
