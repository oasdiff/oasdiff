package diff

import (
	"fmt"
	"mime"
	"strings"
)

type MediaType struct {
	Type       string
	Subtype    string
	Suffix     string
	Parameters map[string]string
}

func ParseMediaType(mediaType string) (*MediaType, error) {
	mediaType, params, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid media type: %s", mediaType)
	}

	result := MediaType{
		Type:       parts[0],
		Subtype:    parts[1],
		Parameters: params,
	}

	subTypeParts := strings.Split(result.Subtype, "+")
	switch len(subTypeParts) {
	case 0:
		return nil, fmt.Errorf("invalid media subtype: %s", mediaType)
	case 1:
		result.Subtype = subTypeParts[0]
		result.Suffix = ""
	case 2:
		result.Subtype = subTypeParts[0]
		result.Suffix = subTypeParts[1]
	default:
		return nil, fmt.Errorf("multiple suffixes not supported: %s", mediaType)
	}

	return &result, nil
}

// IsMediaTypeContained checks if mediaType1 contains mediaType2
// e.g., application/json contains application/problem+json
func IsMediaTypeContained(mediaType1, mediaType2 string) (bool, error) {
	parts1, err := ParseMediaType(mediaType1)
	if err != nil {
		return false, err
	}

	parts2, err := ParseMediaType(mediaType2)
	if err != nil {
		return false, err
	}

	// Types must match (e.g., "application" == "application")
	if parts1.Type != parts2.Type {
		return false, nil
	}

	// is no suffixed, subtypes must match
	if parts1.Suffix == "" && parts2.Suffix == "" {
		return parts1.Subtype == parts2.Subtype, nil
	}

	// if mediaType1 has no suffix, mediaType2 subtype must be the same of mediaType1 subtype
	if parts1.Suffix == "" && parts2.Suffix == parts1.Subtype {
		return true, nil
	}

	// if both have suffixes, they must match and subtype must match
	return parts1.Suffix == parts2.Suffix &&
		parts1.Subtype == parts2.Subtype, nil
}
