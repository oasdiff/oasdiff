package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// EncodingsDiff describes the changes between a pair of sets of encoding objects: https://swagger.io/specification/#encoding-object
type EncodingsDiff struct {
	Added    []string          `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string          `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedEncodings `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// ModifiedEncodings is map of enconding names to their respective diffs
type ModifiedEncodings map[string]*EncodingDiff

// Empty indicates whether a change was found in this element
func (diff *EncodingsDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func newEncodingsDiff() *EncodingsDiff {
	return &EncodingsDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedEncodings{},
	}
}

func getEncodingsDiff(config *Config, state *state, encodings1, encodings2 map[string]*openapi3.Encoding) (*EncodingsDiff, error) {
	diff, err := getEncodingsDiffInternal(config, state, encodings1, encodings2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getEncodingsDiffInternal(config *Config, state *state, encodings1, encodings2 map[string]*openapi3.Encoding) (*EncodingsDiff, error) {

	result := newEncodingsDiff()

	added, deleted, common := keysDiff(encodings1, encodings2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		diff, err := getEncodingDiff(config, state, encodings1[name], encodings2[name])
		if err != nil {
			return nil, err
		}
		if !diff.Empty() {
			result.Modified[name] = diff
		}
	}

	return result, nil
}
