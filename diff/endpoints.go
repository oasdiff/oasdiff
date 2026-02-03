package diff

import "cmp"

// Endpoints is a list of endpoints
type Endpoints []Endpoint

func (endpoints Endpoints) SortFunc(a, b Endpoint) int {
	if c := cmp.Compare(a.Path, b.Path); c != 0 {
		return c
	}
	return cmp.Compare(a.Method, b.Method)
}
