package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

/*
EndpointsDiff is an alternate, simplified view of PathsDiff.
It describes the changes between Endpoints which are a flattened combination of OpenAPI Paths and Operations.

For example, if there's a new path "/test" with method POST then EndpointsDiff will describe this as a new endpoint: POST /test.

Or, if path "/test" was modified to include a new methdod, PUT, then EndpointsDiff will describe this as a new endpoint: PUT /test.
*/
type EndpointsDiff struct {
	Added    Endpoints         `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  Endpoints         `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedEndpoints `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Endpoint is a combination of an HTTP method and a Path
type Endpoint struct {
	Method string `json:"method,omitempty" yaml:"method,omitempty"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *EndpointsDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func newEndpointsDiff() *EndpointsDiff {
	return &EndpointsDiff{
		Added:    Endpoints{},
		Deleted:  Endpoints{},
		Modified: ModifiedEndpoints{},
	}
}

func getEndpointsDiff(config *Config, state *state, paths1, paths2 *openapi3.Paths) (*EndpointsDiff, error) {

	if config.IsExcludeEndpoints() {
		return nil, nil
	}

	if err := filterPaths(config.MatchPath, config.UnmatchPath, config.FilterExtension, paths1, paths2); err != nil {
		return nil, err
	}

	diff, err := getEndpointsDiffInternal(config, state, paths1, paths2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getEndpointsDiffInternal(config *Config, state *state, paths1, paths2 *openapi3.Paths) (*EndpointsDiff, error) {

	result := newEndpointsDiff()

	paths1Mod := rewritePrefix(paths1.Map(), config.PathStripPrefixBase, config.PathPrefixBase)
	paths2Mod := rewritePrefix(paths2.Map(), config.PathStripPrefixRevision, config.PathPrefixRevision)

	addedPaths, deletedPaths, otherPaths := getPathItemsDiff(config, paths1Mod, paths2Mod)

	for path, pathItem := range addedPaths.Map() {
		for method := range pathItem.Operations() {
			result.addAddedPath(path, method)
		}
	}

	for path, pathItem := range deletedPaths.Map() {
		for method := range pathItem.Operations() {
			result.addDeletedPath(path, method)
		}
	}

	for path, pathItemPair := range otherPaths {
		if err := result.addModifiedPaths(config, state, path, pathItemPair); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (diff *EndpointsDiff) addAddedPath(path string, method string) {
	diff.Added = append(diff.Added, Endpoint{
		Method: method,
		Path:   path,
	})
}

func (diff *EndpointsDiff) addDeletedPath(path string, method string) {
	diff.Deleted = append(diff.Deleted, Endpoint{
		Method: method,
		Path:   path,
	})
}

func (diff *EndpointsDiff) addModifiedPaths(config *Config, state *state, path string, pathItemPair *pathItemPair) error {

	pathDiff, err := getPathDiff(config, state, pathItemPair)
	if err != nil {
		return err
	}

	if pathDiff.Empty() || pathDiff.OperationsDiff.Empty() {
		return nil
	}

	for _, method := range pathDiff.OperationsDiff.Added {
		diff.addAddedPath(path, method)
	}

	for _, method := range pathDiff.OperationsDiff.Deleted {
		diff.addDeletedPath(path, method)
	}

	for method, methodDiff := range pathDiff.OperationsDiff.Modified {
		diff.Modified[Endpoint{
			Method: method,
			Path:   path,
		}] = methodDiff
	}

	return nil
}

func (diff *EndpointsDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
