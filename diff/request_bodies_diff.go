package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// RequestBodiesDiff describes the changes between a pair of sets of request body objects: https://swagger.io/specification/#request-body-object
type RequestBodiesDiff struct {
	Added    []string              `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string              `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedRequestBodies `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (requestBodiesDiff *RequestBodiesDiff) Empty() bool {
	if requestBodiesDiff == nil {
		return true
	}

	return len(requestBodiesDiff.Added) == 0 &&
		len(requestBodiesDiff.Deleted) == 0 &&
		len(requestBodiesDiff.Modified) == 0
}

// ModifiedRequestBodies is map of requestBody names to their respective diffs
type ModifiedRequestBodies map[string]*RequestBodyDiff

func newRequestBodiesDiff() *RequestBodiesDiff {
	return &RequestBodiesDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedRequestBodies{},
	}
}

func getRequestBodiesDiff(config *Config, state *state, requestBodies1, requestBodies2 openapi3.RequestBodies) (*RequestBodiesDiff, error) {
	diff, err := getRequestBodiesDiffInternal(config, state, requestBodies1, requestBodies2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getRequestBodiesDiffInternal(config *Config, state *state, requestBodies1, requestBodies2 openapi3.RequestBodies) (*RequestBodiesDiff, error) {

	result := newRequestBodiesDiff()

	added, deleted, common := keysDiff(requestBodies1, requestBodies2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		diff, err := getRequestBodyDiff(config, state, requestBodies1[name], requestBodies2[name])
		if err != nil {
			return nil, err
		}
		if !diff.Empty() {
			result.Modified[name] = diff
		}
	}

	return result, nil
}

func (requestBodiesDiff *RequestBodiesDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(requestBodiesDiff.Added),
		Deleted:  len(requestBodiesDiff.Deleted),
		Modified: len(requestBodiesDiff.Modified),
	}
}
