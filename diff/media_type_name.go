package diff

import (
	"fmt"
	"mime"
	"strings"
)

type MediaTypeName struct {
	Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
	Type       string            `json:"type,omitempty" yaml:"type,omitempty"`
	Subtype    string            `json:"subtype,omitempty" yaml:"subtype,omitempty"`
	Suffix     string            `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

func ParseMediaTypeName(mediaType string) (*MediaTypeName, error) {
	mediaTypeNoParams, params, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse media type '%s': %w", mediaType, err)
	}

	parts := strings.Split(mediaTypeNoParams, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid media type format (missing '/'): %s", mediaTypeNoParams)
	}

	typeName := strings.TrimSpace(parts[0])
	subTypeString := strings.TrimSpace(parts[1])

	result := MediaTypeName{
		Name:       mediaType,
		Type:       typeName,
		Parameters: params,
	}

	subTypeParts := strings.Split(subTypeString, "+")
	result.Subtype = strings.TrimSpace(subTypeParts[0])

	// Handle suffix (only one allowed)
	switch len(subTypeParts) {
	case 1:
		// No suffix
	case 2:
		// One suffix
		result.Suffix = strings.TrimSpace(subTypeParts[1])
		if result.Suffix == "" {
			return nil, fmt.Errorf("invalid media type: empty suffix in '%s'", mediaTypeNoParams)
		}
	default:
		// More than one suffix found
		return nil, fmt.Errorf("multiple suffixes not supported in '%s'", mediaTypeNoParams)
	}

	return &result, nil
}

// IsMediaTypeNameContained checks if mediaType2 is a specialization of mediaType1.
// For example:
// - "application/xml" contains "application/atom+xml" (base type contains suffixed type)
// - "application/json" contains "application/problem+json" (base type contains suffixed type)
// - "application/xml;q=0.9" contains "application/xml;q=0.8" (parameter values are ignored)
func IsMediaTypeNameContained(mediaType1, mediaType2 string) bool {
	parts1, err := ParseMediaTypeName(mediaType1) // Expected/Old
	if err != nil {
		return false
	}
	parts2, err := ParseMediaTypeName(mediaType2) // Actual/New
	if err != nil {
		return false
	}

	// Types must match
	if parts1.Type != parts2.Type {
		return false
	}

	// *** Generalized Refinement Exception ***
	// Check if the original type is a base type (no suffix) and the new type
	// refines it by adding a suffix that matches the original subtype.
	// e.g., "application/xml" contains "application/atom+xml" -> true
	// e.g., "application/json" contains "application/problem+json" -> true
	isPart1BaseType := parts1.Suffix == ""
	doesPart2SuffixMatchPart1Subtype := parts2.Suffix != "" && parts2.Suffix == parts1.Subtype

	if isPart1BaseType && doesPart2SuffixMatchPart1Subtype {
		// Allow refinement from base */subtype to any */*+subtype
		return true
	}

	// *** General Case ***
	// Subtypes must match (if not the refinement exception case)
	if parts1.Subtype != parts2.Subtype {
		return false
	}

	// Suffix Check: Suffixes must be identical if both are present.
	// A suffixed type cannot contain a non-suffixed type, and vice-versa (handled by subtype check mostly).
	if parts1.Suffix != parts2.Suffix {
		return false
	}

	// Types match, subtypes match (or refinement exception), and suffixes match (or follow refinement rule)
	return true
}
