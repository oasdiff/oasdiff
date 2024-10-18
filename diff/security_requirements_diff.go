package diff

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/utils"
)

// SecurityRequirementsDiff describes the changes between a pair of sets of security requirement objects: https://swagger.io/specification/#security-requirement-object
type SecurityRequirementsDiff struct {
	Added    utils.StringList             `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList             `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedSecurityRequirements `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *SecurityRequirementsDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

// ModifiedSecurityRequirements is map of security requirements to their respective diffs
type ModifiedSecurityRequirements map[string]SecurityScopesDiff

func newSecurityRequirementsDiff() *SecurityRequirementsDiff {
	return &SecurityRequirementsDiff{
		Added:    utils.StringList{},
		Deleted:  utils.StringList{},
		Modified: ModifiedSecurityRequirements{},
	}
}

func getSecurityRequirementsDiff(securityRequirements1, securityRequirements2 *openapi3.SecurityRequirements) *SecurityRequirementsDiff {
	diff := getSecurityRequirementsDiffInternal(securityRequirements1, securityRequirements2)

	if diff.Empty() {
		return nil
	}

	return diff
}

func getSecurityRequirementsDiffInternal(securityRequirements1, securityRequirements2 *openapi3.SecurityRequirements) *SecurityRequirementsDiff {

	result := newSecurityRequirementsDiff()

	if securityRequirements1 != nil {
		for _, securityRequirement1 := range *securityRequirements1 {
			if securityRequirement2 := findSecurityRequirement(securityRequirement1, securityRequirements2); securityRequirement2 != nil {
				if securityScopesDiff := getSecurityScopesDiff(securityRequirement1, securityRequirement2); !securityScopesDiff.Empty() {
					result.Modified[getSecurityRequirementID(securityRequirement1)] = securityScopesDiff
				}
			} else {
				result.Deleted = append(result.Deleted, getSecurityRequirementID(securityRequirement1))
			}
		}
	}

	if securityRequirements2 != nil {
		for _, securityRequirement2 := range *securityRequirements2 {
			if securityRequirements1 := findSecurityRequirement(securityRequirement2, securityRequirements1); securityRequirements1 == nil {
				result.Added = append(result.Added, getSecurityRequirementID(securityRequirement2))
			}
		}
	}

	return result
}

func findSecurityRequirement(securityRequirement1 openapi3.SecurityRequirement, securityRequirements2 *openapi3.SecurityRequirements) openapi3.SecurityRequirement {
	if securityRequirements2 == nil {
		return nil
	}

	securitySchemes1 := getSecuritySchemes(securityRequirement1)
	for _, securityRequirement2 := range *securityRequirements2 {
		securitySchemes2 := getSecuritySchemes(securityRequirement2)
		if securitySchemes1.Equals(securitySchemes2) {
			return securityRequirement2
		}
	}
	return nil
}

func getSecuritySchemes(securityRequirement openapi3.SecurityRequirement) utils.StringSet {
	result := utils.StringSet{}
	for name := range securityRequirement {
		result.Add(name)
	}
	return result
}

func getSecurityRequirementID(securityRequirement openapi3.SecurityRequirement) string {
	results := make([]string, len(securityRequirement))
	i := 0
	for name := range securityRequirement {
		results[i] = name
		i++
	}
	return strings.Join(results, " AND ")
}

func (diff *SecurityRequirementsDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
