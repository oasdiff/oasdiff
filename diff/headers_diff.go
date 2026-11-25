package diff

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// HeadersDiff describes the changes between a pair of sets of header objects: https://swagger.io/specification/#header-object
type HeadersDiff struct {
	Added    []string        `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string        `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedHeaders `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (headersDiff *HeadersDiff) Empty() bool {
	if headersDiff == nil {
		return true
	}

	return len(headersDiff.Added) == 0 &&
		len(headersDiff.Deleted) == 0 &&
		len(headersDiff.Modified) == 0
}

// ModifiedHeaders is map of header names to their respective diffs
type ModifiedHeaders map[string]*HeaderDiff

func newHeadersDiff() *HeadersDiff {
	return &HeadersDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedHeaders{},
	}
}

func getHeadersDiff(config *Config, state *state, headers1, headers2 openapi3.Headers) (*HeadersDiff, error) {
	diff, err := getHeadersDiffInternal(config, state, headers1, headers2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getHeadersDiffInternal(config *Config, state *state, headers1, headers2 openapi3.Headers) (*HeadersDiff, error) {

	result := newHeadersDiff()

	added, deleted, common := keysDiff(headers1, headers2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		if err := result.addModified(config, state, name, headers1[name], headers2[name]); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// addModified compares two headers and records their diff, if any, under name.
func (headersDiff *HeadersDiff) addModified(config *Config, state *state, name string, header1, header2 *openapi3.HeaderRef) error {
	value1, err := derefHeader(header1)
	if err != nil {
		return err
	}

	value2, err := derefHeader(header2)
	if err != nil {
		return err
	}

	diff, err := getHeaderDiff(config, state, value1, value2)
	if err != nil {
		return err
	}

	if !diff.Empty() {
		if headersDiff.Modified == nil {
			headersDiff.Modified = ModifiedHeaders{}
		}
		headersDiff.Modified[name] = diff
	}
	return nil
}

// getHeadersDiffIgnoringCase returns headersDiff with each deleted header that
// was added under a name differing only in case compared as one header. It
// returns headersDiff itself when there is none.
func getHeadersDiffIgnoringCase(config *Config, state *state, headers1, headers2 openapi3.Headers, headersDiff *HeadersDiff) (*HeadersDiff, error) {
	if headersDiff == nil {
		return nil, nil
	}

	result := &HeadersDiff{
		Added:    slices.Clone(headersDiff.Added),
		Deleted:  []string{},
		Modified: maps.Clone(headersDiff.Modified),
	}

	for _, name1 := range headersDiff.Deleted {
		i := slices.IndexFunc(result.Added, func(name2 string) bool {
			return strings.EqualFold(name1, name2)
		})
		if i == -1 {
			result.Deleted = append(result.Deleted, name1)
			continue
		}
		name2 := result.Added[i]
		result.Added = slices.Delete(result.Added, i, i+1)

		if err := result.addModified(config, state, name1, headers1[name1], headers2[name2]); err != nil {
			return nil, err
		}
	}

	if len(result.Deleted) == len(headersDiff.Deleted) {
		return headersDiff, nil
	}

	if result.Empty() {
		return nil, nil
	}

	return result, nil
}

func derefHeader(ref *openapi3.HeaderRef) (*openapi3.Header, error) {

	if ref == nil || ref.Value == nil {
		return nil, fmt.Errorf("header reference is nil")
	}

	return ref.Value, nil
}

func (headersDiff *HeadersDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(headersDiff.Added),
		Deleted:  len(headersDiff.Deleted),
		Modified: len(headersDiff.Modified),
	}
}
