package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// SecuritySchemesDiff describes the changes between a pair of sets of security scheme objects: https://swagger.io/specification/#security-scheme-object
type SecuritySchemesDiff struct {
	Added    []string                 `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string                 `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedSecuritySchemes  `json:"modified,omitempty" yaml:"modified,omitempty"`
	Base     openapi3.SecuritySchemes `json:"-" yaml:"-"`
	Revision openapi3.SecuritySchemes `json:"-" yaml:"-"`
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
		Added:    []string{},
		Deleted:  []string{},
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
	result.Base = securitySchemes1
	result.Revision = securitySchemes2

	added, deleted, common := keysDiff(securitySchemes1, securitySchemes2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		value1, err := derefSecurityScheme(securitySchemes1[name])
		if err != nil {
			return nil, err
		}
		value2, err := derefSecurityScheme(securitySchemes2[name])
		if err != nil {
			return nil, err
		}
		diff, err := getSecuritySchemeDiff(config, value1, value2)
		if err != nil {
			return nil, err
		}
		if !diff.Empty() {
			result.Modified[name] = diff
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
