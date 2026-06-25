package diff

import (
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

// SecurityRequirementsDiff describes the changes between a pair of sets of security requirement objects: https://swagger.io/specification/#security-requirement-object
type SecurityRequirementsDiff struct {
	Added    []string                     `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string                     `json:"deleted,omitempty" yaml:"deleted,omitempty"`
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
		Added:    []string{},
		Deleted:  []string{},
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

	// Security requirements are an unordered set of OR-alternatives, and several
	// alternatives may share a scheme (e.g. `- petstore_auth: [read:pets]` /
	// `- petstore_auth: [write:pets]`), so they cannot be matched by scheme name alone.
	reqs1 := derefSecurityRequirements(securityRequirements1)
	reqs2 := derefSecurityRequirements(securityRequirements2)
	matched1 := make([]bool, len(reqs1))
	matched2 := make([]bool, len(reqs2))

	// Match unchanged alternatives (same schemes and scopes) so they don't participate in scope diffing.
	for i := range reqs1 {
		if j := findSecurityRequirementIndex(reqs1[i], reqs2, matched2, true); j >= 0 {
			matched1[i] = true
			matched2[j] = true
		}
	}

	// Report a leftover as a scope modification only when its scheme set is unambiguous:
	// exactly one unmatched alternative with that scheme set on each side. Otherwise any
	// pairing is arbitrary, so the alternatives fall through to deleted/added below.
	for i := range reqs1 {
		if matched1[i] || !uniqueUnmatchedBySchemes(reqs1, matched1, getSecuritySchemes(reqs1[i])) {
			continue
		}
		j := findSecurityRequirementIndex(reqs1[i], reqs2, matched2, false)
		if j < 0 || !uniqueUnmatchedBySchemes(reqs2, matched2, getSecuritySchemes(reqs2[j])) {
			continue
		}
		matched1[i] = true
		matched2[j] = true
		if securityScopesDiff := getSecurityScopesDiff(reqs1[i], reqs2[j]); !securityScopesDiff.Empty() {
			result.Modified[getSecurityRequirementID(reqs1[i])] = securityScopesDiff
		}
	}

	for i := range reqs1 {
		if !matched1[i] {
			result.Deleted = append(result.Deleted, getSecurityRequirementID(reqs1[i]))
		}
	}
	for j := range reqs2 {
		if !matched2[j] {
			result.Added = append(result.Added, getSecurityRequirementID(reqs2[j]))
		}
	}

	return result
}

func derefSecurityRequirements(securityRequirements *openapi3.SecurityRequirements) []openapi3.SecurityRequirement {
	if securityRequirements == nil {
		return nil
	}
	return *securityRequirements
}

// findSecurityRequirementIndex returns the index of the first unmatched
// requirement in candidates that matches securityRequirement, or -1 if none.
// When exact is true the scopes must match too; otherwise only the set of
// scheme names must match.
func findSecurityRequirementIndex(securityRequirement openapi3.SecurityRequirement, candidates []openapi3.SecurityRequirement, matched []bool, exact bool) int {
	schemes := getSecuritySchemes(securityRequirement)
	for j, candidate := range candidates {
		if matched[j] {
			continue
		}
		if !schemes.Equals(getSecuritySchemes(candidate)) {
			continue
		}
		if exact && !getSecurityScopesDiff(securityRequirement, candidate).Empty() {
			continue
		}
		return j
	}
	return -1
}

// uniqueUnmatchedBySchemes reports whether exactly one unmatched requirement in
// reqs has the given set of scheme names.
func uniqueUnmatchedBySchemes(reqs []openapi3.SecurityRequirement, matched []bool, schemes utils.StringSet) bool {
	count := 0
	for i := range reqs {
		if matched[i] {
			continue
		}
		if schemes.Equals(getSecuritySchemes(reqs[i])) {
			count++
			if count > 1 {
				return false
			}
		}
	}
	return count == 1
}

func getSecuritySchemes(securityRequirement openapi3.SecurityRequirement) utils.StringSet {
	result := utils.StringSet{}
	for name := range securityRequirement {
		result.Add(name)
	}
	return result
}

func getSecurityRequirementID(securityRequirement openapi3.SecurityRequirement) string {
	return strings.Join(slices.Sorted(maps.Keys(securityRequirement)), " AND ")
}

func (diff *SecurityRequirementsDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
