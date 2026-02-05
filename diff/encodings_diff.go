package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// EncodingsDiff describes the changes between a pair of sets of encoding objects: https://swagger.io/specification/#encoding-object
type EncodingsDiff struct {
	Added    []string  `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string  `json:"deleted,omitempty" yaml:"deleted,omitempty"`
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

	for name1, encoding1 := range encodings1 {
		if encoding2, ok := encodings2[name1]; ok {
			diff, err := getEncodingDiff(config, state, encoding1, encoding2)
			if err != nil {
				return nil, err
			}
			if !diff.Empty() {
				result.Modified[name1] = diff
			}
		} else {
			result.Deleted = append(result.Deleted, name1)
		}
	}

	for name2 := range encodings2 {
		if _, ok := encodings1[name2]; !ok {
			result.Added = append(result.Added, name2)
		}
	}

	return result, nil
}
