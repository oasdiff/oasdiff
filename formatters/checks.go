package formatters

import "cmp"

// Check is one rule in a `oasdiff checks` listing.
//
// Every field except Id is omitempty: the changelog and breaking-change rules
// populate all of them (TestChecks_AllFieldsPopulated pins that), so their
// output is unchanged, while a listing that only has IDs to report
// (`oasdiff checks validate`, where a rule's text and severity are determined
// at runtime) renders as ids alone instead of rows of empty strings.
type Check struct {
	Id          string `json:"id" yaml:"id"`
	Level       string `json:"level,omitempty" yaml:"level,omitempty"`
	Direction   string `json:"direction,omitempty" yaml:"direction,omitempty"`
	Area        string `json:"area,omitempty" yaml:"area,omitempty"`
	Kind        string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Action      string `json:"action,omitempty" yaml:"action,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Mitigation  string `json:"mitigation,omitempty" yaml:"mitigation,omitempty"`
}

type Checks []Check

func (checks Checks) SortFunc(a, b Check) int {
	return cmp.Compare(a.Id, b.Id)
}
