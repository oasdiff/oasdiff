package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// LicenseDiff describes the changes between a pair of license objects: https://swagger.io/specification/#license-object
type LicenseDiff struct {
	Added          bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted        bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	NameDiff       *ValueDiff      `json:"name,omitempty" yaml:"name,omitempty"`
	URLDiff        *ValueDiff      `json:"url,omitempty" yaml:"url,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *LicenseDiff) Empty() bool {
	return diff == nil || *diff == LicenseDiff{}
}

func getLicenseDiff(config *Config, license1, license2 *openapi3.License) (*LicenseDiff, error) {
	diff, err := getLicenseDiffInternal(config, license1, license2)

	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getLicenseDiffInternal(config *Config, license1, license2 *openapi3.License) (*LicenseDiff, error) {

	result := LicenseDiff{}
	var err error

	if license1 == nil && license2 == nil {
		return &result, nil
	}

	if license1 == nil && license2 != nil {
		result.Added = true
		return &result, nil
	}

	if license1 != nil && license2 == nil {
		result.Deleted = true
		return &result, nil
	}

	result.ExtensionsDiff, err = getExtensionsDiff(config, license1.Extensions, license2.Extensions)
	if err != nil {
		return nil, err
	}
	result.NameDiff = getValueDiff(license1.Name, license2.Name)
	result.URLDiff = getValueDiff(license1.URL, license2.URL)

	return &result, nil
}
