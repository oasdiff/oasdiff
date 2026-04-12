package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// InfoDiff describes the changes between a pair of info objects: https://swagger.io/specification/#info-object
type InfoDiff struct {
	Added              bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted            bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff     *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	TitleDiff          *ValueDiff      `json:"title,omitempty" yaml:"title,omitempty"`
	DescriptionDiff    *ValueDiff      `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfServiceDiff *ValueDiff      `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	ContactDiff        *ContactDiff    `json:"contact,omitempty" yaml:"contact,omitempty"`
	LicenseDiff        *LicenseDiff    `json:"license,omitempty" yaml:"license,omitempty"`
	VersionDiff        *ValueDiff      `json:"version,omitempty" yaml:"version,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *InfoDiff) Empty() bool {
	return diff == nil || *diff == InfoDiff{}
}

func getInfoDiff(config *Config, info1, info2 *openapi3.Info) (*InfoDiff, error) {
	diff, err := getInfoDiffInternal(config, info1, info2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getInfoDiffInternal(config *Config, info1, info2 *openapi3.Info) (*InfoDiff, error) {

	if info1 == nil && info2 == nil {
		return nil, nil
	}

	if info1 == nil && info2 != nil {
		return &InfoDiff{
			Added: true,
		}, nil
	}

	if info1 != nil && info2 == nil {
		return &InfoDiff{
			Deleted: true,
		}, nil
	}

	extensionsDiff, err := getExtensionsDiff(config, info1.Extensions, info2.Extensions)
	if err != nil {
		return nil, err
	}
	licenseDiff, err := getLicenseDiff(config, info1.License, info2.License)
	if err != nil {
		return nil, err
	}
	contactDiff, err := getContactDiff(config, info1.Contact, info2.Contact)
	if err != nil {
		return nil, err
	}

	return &InfoDiff{
		ExtensionsDiff:     extensionsDiff,
		TitleDiff:          getValueDiffConditional(config.IsExcludeTitle(), info1.Title, info2.Title),
		DescriptionDiff:    getValueDiffConditional(config.IsExcludeDescription(), info1.Description, info2.Description),
		TermsOfServiceDiff: getValueDiff(info1.TermsOfService, info2.TermsOfService),
		ContactDiff:        contactDiff,
		LicenseDiff:        licenseDiff,
		VersionDiff:        getValueDiff(info1.Version, info2.Version),
	}, nil
}
