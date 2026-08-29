package formatters

import "cmp"

// Check is one rule in a `oasdiff checks` listing; the fields render the
// rule model (see checker/rules.Rule).
//
// Every field except Id is omitempty. The changelog and breaking-change
// rules populate all of them, so their output is unchanged; the validate
// rules have an id, a description and a level but no direction, area,
// kind, actions or effect, and those are omitted rather than rendered as
// empty strings.
type Check struct {
	Id          string   `json:"id" yaml:"id"`
	Level       string   `json:"level,omitempty" yaml:"level,omitempty"`
	Direction   string   `json:"direction,omitempty" yaml:"direction,omitempty"`
	Area        string   `json:"area,omitempty" yaml:"area,omitempty"`
	Kind        string   `json:"kind,omitempty" yaml:"kind,omitempty"`
	Actions     []string `json:"actions,omitempty" yaml:"actions,omitempty"`
	Effect      string   `json:"effect,omitempty" yaml:"effect,omitempty"`
	Locations   []string `json:"locations,omitempty" yaml:"locations,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Mitigation  string   `json:"mitigation,omitempty" yaml:"mitigation,omitempty"`
}

type Checks []Check

func (checks Checks) SortFunc(a, b Check) int {
	return cmp.Compare(a.Id, b.Id)
}
