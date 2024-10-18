package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/utils"
)

// SecuritySchemesDiff describes the changes between a pair of sets of security scheme objects: https://swagger.io/specification/#security-scheme-object
type SecuritySchemesDiff struct {
	Added    utils.StringList        `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList        `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedSecuritySchemes `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *SecuritySchemesDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

// ModifiedSecuritySchemes is map of security schemes to their respective diffs
type ModifiedSecuritySchemes map[string]*SecuritySchemeDiff

func newSecuritySchemesDiff() *SecuritySchemesDiff {
	return &SecuritySchemesDiff{
		Added:    utils.StringList{},
		Deleted:  utils.StringList{},
		Modified: ModifiedSecuritySchemes{},
	}
}

func getSecuritySchemesDiff(config *Config, securitySchemes1, securitySchemes2 openapi3.SecuritySchemes) (*SecuritySchemesDiff, error) {
	diff, err := getSecuritySchemesDiffInternal(config, securitySchemes1, securitySchemes2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getSecuritySchemesDiffInternal(config *Config, securitySchemes1, securitySchemes2 openapi3.SecuritySchemes) (*SecuritySchemesDiff, error) {

	result := newSecuritySchemesDiff()

	for name1, ref1 := range securitySchemes1 {
		if ref2, ok := securitySchemes2[name1]; ok {
			value1, err := derefSecurityScheme(ref1)
			if err != nil {
				return nil, err
			}
			value2, err := derefSecurityScheme(ref2)
			if err != nil {
				return nil, err
			}
			diff, err := getSecuritySchemeDiff(config, value1, value2)
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

	for value2 := range securitySchemes2 {
		if _, ok := securitySchemes1[value2]; !ok {
			result.Added = append(result.Added, value2)
		}
	}

	return result, nil
}

func derefSecurityScheme(ref *openapi3.SecuritySchemeRef) (*openapi3.SecurityScheme, error) {

	if ref == nil || ref.Value == nil {
		return nil, fmt.Errorf("security scheme reference is nil")
	}

	return ref.Value, nil
}

func (diff *SecuritySchemesDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
