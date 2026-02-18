package diff

import (
	"maps"
	"slices"
)

// ModifiedEndpoints is a map of endpoints to their respective diffs
type ModifiedEndpoints map[Endpoint]*MethodDiff

// ToEndpoints returns the modified endpoints
func (modifiedEndpoints ModifiedEndpoints) ToEndpoints() Endpoints {
	return Endpoints(slices.Collect(maps.Keys(modifiedEndpoints)))
}
