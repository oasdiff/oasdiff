package formatters

import "cmp"

type Check struct {
	Id          string `json:"id" yaml:"id"`
	Level       string `json:"level" yaml:"level"`
	Description string `json:"description" yaml:"description"`
}

type Checks []Check

func (checks Checks) SortFunc(a, b Check) int {
	return cmp.Compare(a.Id, b.Id)
}
