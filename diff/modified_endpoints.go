package diff

import (
	"maps"
	"slices"

	"go.yaml.in/yaml/v3"
)

// ModifiedEndpoints is a map of endpoints to their respective diffs
type ModifiedEndpoints map[Endpoint]*MethodDiff

// ToEndpoints returns the modified endpoints, sorted by path and method
func (modifiedEndpoints ModifiedEndpoints) ToEndpoints() Endpoints {
	endpoints := Endpoints(slices.Collect(maps.Keys(modifiedEndpoints)))
	slices.SortFunc(endpoints, endpoints.SortFunc)
	return endpoints
}

// MarshalYAML writes the endpoints sorted by path and method. yaml sorts
// string keys itself but leaves struct keys in map order.
func (modifiedEndpoints ModifiedEndpoints) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, endpoint := range modifiedEndpoints.ToEndpoints() {
		var key, value yaml.Node
		if err := key.Encode(endpoint); err != nil {
			return nil, err
		}
		if err := value.Encode(modifiedEndpoints[endpoint]); err != nil {
			return nil, err
		}
		node.Content = append(node.Content, &key, &value)
	}
	return node, nil
}
