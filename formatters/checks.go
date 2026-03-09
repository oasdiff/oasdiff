package formatters

import "cmp"

type Check struct {
	Id          string `json:"id" yaml:"id"`
	Level       string `json:"level" yaml:"level"`
	Direction   string `json:"direction" yaml:"direction"`
	Location    string `json:"location" yaml:"location"`
	Action      string `json:"action" yaml:"action"`
	Description string `json:"description" yaml:"description"`
	Mitigation  string `json:"mitigation,omitempty" yaml:"mitigation,omitempty"`
}

type Checks []Check

func (checks Checks) SortFunc(a, b Check) int {
	return cmp.Compare(a.Id, b.Id)
}
