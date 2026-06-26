package diff

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

// SecurityRequirementsDiff describes the changes between a pair of lists of security requirement objects: https://swagger.io/specification/#security-requirement-object
//
// Semantics, which drive the modeling below:
//   - the list is an OR: a request is authorized if it satisfies any one item;
//   - the schemes within one item are AND-ed: all of them must be satisfied;
//   - the scopes within a scheme are AND-ed too.
//
// So an OR of scopes for a single scheme can only be written by repeating the
// scheme across items (- petstore_auth: [read] / - petstore_auth: [write]), and
// a single item may carry several schemes AND-ed together (both an oauth and an
// apiKey key).
//
// An alternative therefore has no identity of its own to key on: its scheme
// names aren't unique (the repeated-scheme case above) and there may be several
// of them. The scheme name is only a reference into components.securitySchemes,
// an author-chosen label; the scheme's actual meaning (type, flows) is diffed
// separately in SecuritySchemesDiff. So alternatives cannot be keyed by a
// string; like SubschemasDiff, they are identified by index and carried as
// structured values.
type SecurityRequirementsDiff struct {
	Added    SecurityAlternatives         `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  SecurityAlternatives         `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedSecurityRequirements `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// SecurityAlternative is one alternative (one security requirement object) in a
// security list, identified by its index plus its schemes and their scopes.
type SecurityAlternative struct {
	Index   int                 `json:"index" yaml:"index"`     // zero-based index in the security list
	Schemes map[string][]string `json:"schemes" yaml:"schemes"` // scheme name -> scopes
}

// SecurityAlternatives is a list of security alternatives.
type SecurityAlternatives []SecurityAlternative

// String renders an alternative by its schemes and scopes. The index is kept in
// the structured data but left out of the human-readable form, since it is
// positional rather than a meaningful identity. Schemes and scopes are sorted so
// the output is deterministic. A scheme with no scopes (API key, bearer) shows
// as its bare name; an empty requirement (`{}`, this alternative requires no
// authentication) shows as "{}".
func (a SecurityAlternative) String() string {
	if len(a.Schemes) == 0 {
		return "{}"
	}
	schemes := slices.Sorted(maps.Keys(a.Schemes))
	parts := make([]string, len(schemes))
	for i, scheme := range schemes {
		scopes := slices.Clone(a.Schemes[scheme])
		if len(scopes) == 0 {
			parts[i] = scheme
			continue
		}
		slices.Sort(scopes)
		parts[i] = fmt.Sprintf("%s: [%s]", scheme, strings.Join(scopes, ", "))
	}
	return strings.Join(parts, " AND ")
}

// SchemeNames returns the alternative's scheme names, sorted and AND-joined
// (e.g. "apiKey AND oauth"). It labels a scope modification, where the changed
// scopes are reported separately, so it omits the scopes that String() carries.
func (a SecurityAlternative) SchemeNames() string {
	return strings.Join(slices.Sorted(maps.Keys(a.Schemes)), " AND ")
}

func newSecurityAlternative(index int, securityRequirement openapi3.SecurityRequirement) SecurityAlternative {
	return SecurityAlternative{
		Index:   index,
		Schemes: maps.Clone(securityRequirement),
	}
}

// ModifiedSecurityRequirements is a list of alternatives whose scopes changed.
// Like ModifiedSubschemas, it is modeled as a slice rather than a map to avoid a
// composite key.
type ModifiedSecurityRequirements []*ModifiedSecurityRequirement

// ModifiedSecurityRequirement is an alternative whose scopes changed, with its
// identity in base and revision and the per-scheme scope diff.
type ModifiedSecurityRequirement struct {
	Base     SecurityAlternative `json:"base" yaml:"base"`
	Revision SecurityAlternative `json:"revision" yaml:"revision"`
	Scopes   SecurityScopesDiff  `json:"scopes" yaml:"scopes"`
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

func newSecurityRequirementsDiff() *SecurityRequirementsDiff {
	return &SecurityRequirementsDiff{
		Added:    SecurityAlternatives{},
		Deleted:  SecurityAlternatives{},
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

	// A security requirements list is an unordered set of OR-alternatives, and
	// several alternatives may legitimately share a scheme while differing only
	// in scopes: because scopes within one requirement are AND-ed, "scope read
	// OR scope write" can only be written as `- petstore_auth: [read]` /
	// `- petstore_auth: [write]`. Matching alternatives by scheme name alone
	// collapses those together and invents a phantom scope add/remove, even when
	// a spec is compared against itself (#1041). So match alternatives as whole
	// units (schemes AND scopes) first, and only fall back to a scope diff when
	// the leftover's scheme set is unambiguous.
	reqs1 := derefSecurityRequirements(securityRequirements1)
	reqs2 := derefSecurityRequirements(securityRequirements2)
	matched1 := make([]bool, len(reqs1))
	matched2 := make([]bool, len(reqs2))

	// Pass 1: cancel out identical alternatives (same schemes and same scopes),
	// so they take no part in scope diffing.
	for i := range reqs1 {
		if j := findSecurityRequirement(reqs1[i], reqs2, matched2, true); j >= 0 {
			matched1[i], matched2[j] = true, true
		}
	}

	// Pass 2: report a leftover as a scope modification only when its scheme set
	// is unambiguous, i.e. exactly one unmatched alternative carries that scheme
	// set on each side. Otherwise any pairing would be arbitrary, so leave those
	// to be reported as a deleted alternative plus an added one in pass 3.
	for i := range reqs1 {
		if matched1[i] || !uniqueUnmatchedBySchemes(reqs1, matched1, getSecuritySchemes(reqs1[i])) {
			continue
		}
		j := findSecurityRequirement(reqs1[i], reqs2, matched2, false)
		if j < 0 || !uniqueUnmatchedBySchemes(reqs2, matched2, getSecuritySchemes(reqs2[j])) {
			continue
		}
		matched1[i], matched2[j] = true, true
		if scopesDiff := getSecurityScopesDiff(reqs1[i], reqs2[j]); !scopesDiff.Empty() {
			result.Modified = append(result.Modified, &ModifiedSecurityRequirement{
				Base:     newSecurityAlternative(i, reqs1[i]),
				Revision: newSecurityAlternative(j, reqs2[j]),
				Scopes:   scopesDiff,
			})
		}
	}

	// Pass 3: whatever is still unmatched was genuinely deleted or added.
	for i := range reqs1 {
		if !matched1[i] {
			result.Deleted = append(result.Deleted, newSecurityAlternative(i, reqs1[i]))
		}
	}
	for j := range reqs2 {
		if !matched2[j] {
			result.Added = append(result.Added, newSecurityAlternative(j, reqs2[j]))
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

// findSecurityRequirement returns the index of the first unmatched alternative
// in candidates whose set of scheme names equals that of securityRequirement,
// or -1 if there is none. When exact is true the scopes must match too, so a
// match means the two alternatives are identical.
func findSecurityRequirement(securityRequirement openapi3.SecurityRequirement, candidates []openapi3.SecurityRequirement, matched []bool, exact bool) int {
	schemes := getSecuritySchemes(securityRequirement)
	for j, candidate := range candidates {
		if matched[j] {
			continue
		}
		if !schemes.Equals(getSecuritySchemes(candidate)) {
			continue
		}
		// getSecurityScopesDiff iterates securityRequirement's schemes; the
		// schemes.Equals check above guarantees both sides share the same scheme
		// names, so an empty diff here means the scopes are identical too.
		if exact && !getSecurityScopesDiff(securityRequirement, candidate).Empty() {
			continue
		}
		return j
	}
	return -1
}

// uniqueUnmatchedBySchemes reports whether exactly one unmatched alternative in
// reqs carries the given set of scheme names.
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

func (diff *SecurityRequirementsDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
