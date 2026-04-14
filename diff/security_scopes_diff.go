package diff

import "github.com/getkin/kin-openapi/openapi3"

// SecurityScopesDiff is a map of security schemes to their respective scope diffs
type SecurityScopesDiff map[string]*StringsDiff

// Empty indicates whether a change was found in this element
func (diff SecurityScopesDiff) Empty() bool {
	return len(diff) == 0
}

func getSecurityScopesDiff(securityRequirement1, securityRequirements2 openapi3.SecurityRequirement) SecurityScopesDiff {
	result := SecurityScopesDiff{}
	for scheme1, scopes1 := range securityRequirement1 {
		if scopes2, ok := securityRequirements2[scheme1]; ok {
			if scopeDiff := getStringsDiff(scopes1, scopes2); !scopeDiff.Empty() {
				result[scheme1] = getStringsDiff(scopes1, scopes2)
			}
		}
	}
	return result
}
