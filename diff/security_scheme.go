package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// SecuritySchemeDiff describes the changes between a pair of security scheme objects: https://swagger.io/specification/#security-scheme-object
type SecuritySchemeDiff struct {
	ExtensionsDiff       *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	TypeDiff             *ValueDiff      `json:"type,omitempty" yaml:"type,omitempty"`
	DescriptionDiff      *ValueDiff      `json:"description,omitempty" yaml:"description,omitempty"`
	NameDiff             *ValueDiff      `json:"name,omitempty" yaml:"name,omitempty"`
	InDiff               *ValueDiff      `json:"in,omitempty" yaml:"in,omitempty"`
	SchemeDiff           *ValueDiff      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormatDiff     *ValueDiff      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	OAuthFlowsDiff       *OAuthFlowsDiff `json:"OAuthFlows,omitempty" yaml:"OAuthFlows,omitempty"`
	OpenIDConnectURLDiff *ValueDiff      `json:"openIDConnectURL,omitempty" yaml:"openIDConnectURL,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *SecuritySchemeDiff) Empty() bool {
	return diff == nil || *diff == SecuritySchemeDiff{}
}

func getSecuritySchemeDiff(config *Config, scheme1, scheme2 *openapi3.SecurityScheme) (*SecuritySchemeDiff, error) {
	diff, err := getSecuritySchemeDiffInternal(config, scheme1, scheme2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getSecuritySchemeDiffInternal(config *Config, scheme1, scheme2 *openapi3.SecurityScheme) (*SecuritySchemeDiff, error) {
	result := SecuritySchemeDiff{}
	var err error

	result.ExtensionsDiff, err = getExtensionsDiff(config, scheme1.Extensions, scheme2.Extensions)
	if err != nil {
		return nil, err
	}

	result.TypeDiff = getValueDiff(scheme1.Type, scheme2.Type)
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), scheme1.Description, scheme2.Description)
	result.NameDiff = getValueDiff(scheme1.Name, scheme2.Name)
	result.InDiff = getValueDiff(scheme1.In, scheme2.In)
	result.SchemeDiff = getValueDiff(scheme1.Scheme, scheme2.Scheme)
	result.BearerFormatDiff = getValueDiff(scheme1.BearerFormat, scheme2.BearerFormat)
	result.OAuthFlowsDiff, err = getOAuthFlowsDiff(config, scheme1.Flows, scheme2.Flows)
	if err != nil {
		return nil, err
	}
	result.OpenIDConnectURLDiff = getValueDiff(scheme1.OpenIdConnectUrl, scheme2.OpenIdConnectUrl)

	return &result, nil
}
