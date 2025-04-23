package diff

import (
	"fmt"
	"mime"
	"strings"

	"github.com/oasdiff/oasdiff/utils"
)

type MediaTypeName struct {
	Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
	Type       string            `json:"type,omitempty" yaml:"type,omitempty"`
	Subtype    string            `json:"subtype,omitempty" yaml:"subtype,omitempty"`
	Suffixes   utils.StringList  `json:"suffixes,omitempty" yaml:"suffixes,omitempty"`
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
	if typeName == "" {
		return nil, fmt.Errorf("invalid media type: empty type in '%s'", mediaTypeNoParams)
	}

	subTypeString := strings.TrimSpace(parts[1])
	if subTypeString == "" {
		return nil, fmt.Errorf("invalid media type: empty subtype in '%s'", mediaTypeNoParams)
	}

	result := MediaTypeName{
		Name:       mediaType,
		Type:       typeName,
		Parameters: params,
		Suffixes:   utils.StringList{}, // Use utils.StringList type
	}

	subTypeParts := strings.Split(subTypeString, "+")
	result.Subtype = strings.TrimSpace(subTypeParts[0])
	if result.Subtype == "" {
		return nil, fmt.Errorf("invalid media type: empty base subtype in '%s'", mediaTypeNoParams)
	}

	if len(subTypeParts) > 1 {
		suffixes := make(utils.StringList, 0, len(subTypeParts)-1)
		for _, suffix := range subTypeParts[1:] {
			trimmedSuffix := strings.TrimSpace(suffix)
			if trimmedSuffix == "" {
				return nil, fmt.Errorf("invalid media type: empty suffix in '%s'", mediaTypeNoParams)
			}
			suffixes = append(suffixes, trimmedSuffix)
		}
		result.Suffixes = suffixes
	}

	return &result, nil
}

// IsMediaTypeNameContained checks if mediaType2 is a specific sub-type of mediaType1
// Examples:
// - "application/json" contains "application/problem+json" -> true (JSON exception)
// - "application/problem+json" contains "application/json" -> false
// - "image/png+json" contains "image/png+json+xml" -> true (Suffix refinement)
// - "image/png+json+xml" contains "image/png+json" -> false
func IsMediaTypeNameContained(mediaType1, mediaType2 string) bool { // Can mediaType2 be safely used where mediaType1 was expected?
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
	// Check if the original type is a base type (no suffixes) and the new type
	// refines it by adding one or more suffixes where the *last* suffix
	// matches the original subtype.
	// e.g., "application/xml" contains "application/atom+xml" -> true
	// e.g., "application/json" contains "application/problem+json" -> true
	isPart1BaseType := len(parts1.Suffixes) == 0
	lastSuffix2 := ""
	if len(parts2.Suffixes) > 0 {
		lastSuffix2 = parts2.Suffixes[len(parts2.Suffixes)-1]
	}

	if isPart1BaseType && lastSuffix2 == parts1.Subtype {
		// Allow refinement from base */subtype to any */*...+subtype
		return true
	}

	// *** General Case ***
	// Subtypes must match (if not the refinement exception case)
	if parts1.Subtype != parts2.Subtype {
		return false
	}

	// Suffix Check: The new type's suffixes (parts2) must start with the old type's suffixes (parts1).
	// The new type can have additional suffixes appended.
	len1 := len(parts1.Suffixes)
	len2 := len(parts2.Suffixes)
	if len2 < len1 { // New type cannot have fewer suffixes than the old type
		return false
	}

	// Compare the prefix (left-to-right)
	for i := 0; i < len1; i++ {
		if parts1.Suffixes[i] != parts2.Suffixes[i] {
			return false // Old suffixes must be a prefix of new suffixes
		}
	}

	// Types match, subtypes match (or refinement exception), and old suffixes are a prefix of new suffixes
	return true
}
