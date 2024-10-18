package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ContactDiff describes the changes between a pair of contact objects: https://swagger.io/specification/#contact-object
type ContactDiff struct {
	Added          bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted        bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	NameDiff       *ValueDiff      `json:"name,omitempty" yaml:"name,omitempty"`
	URLDiff        *ValueDiff      `json:"url,omitempty" yaml:"url,omitempty"`
	EmailDiff      *ValueDiff      `json:"email,omitempty" yaml:"email,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *ContactDiff) Empty() bool {
	return diff == nil || *diff == ContactDiff{}
}

func getContactDiff(config *Config, contact1, contact2 *openapi3.Contact) (*ContactDiff, error) {
	diff, err := getContactDiffInternal(config, contact1, contact2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getContactDiffInternal(config *Config, contact1, contact2 *openapi3.Contact) (*ContactDiff, error) {

	result := ContactDiff{}
	var err error

	if contact1 == nil && contact2 == nil {
		return &result, nil
	}

	if contact1 == nil && contact2 != nil {
		result.Added = true
		return &result, nil
	}

	if contact1 != nil && contact2 == nil {
		result.Deleted = true
		return &result, nil
	}

	result.ExtensionsDiff, err = getExtensionsDiff(config, contact1.Extensions, contact2.Extensions)
	if err != nil {
		return nil, err
	}
	result.NameDiff = getValueDiff(contact1.Name, contact2.Name)
	result.URLDiff = getValueDiff(contact1.URL, contact2.URL)
	result.EmailDiff = getValueDiff(contact1.Email, contact2.Email)

	return &result, nil
}
