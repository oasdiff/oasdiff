package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// OAuthFlowDiff describes the changes between a pair of oauth flow objects: https://swagger.io/specification/#oauth-flow-object
type OAuthFlowDiff struct {
	Added                bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted              bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff       *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	AuthorizationURLDiff *ValueDiff      `json:"authorizationURL,omitempty" yaml:"authorizationURL,omitempty"`
	TokenURLDiff         *ValueDiff      `json:"tokenURL,omitempty" yaml:"tokenURL,omitempty"`
	RefreshURLDiff       *ValueDiff      `json:"refresh,omitempty" yaml:"refresh,omitempty"`
	ScopesDiff           *StringMapDiff  `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *OAuthFlowDiff) Empty() bool {
	return diff == nil || *diff == OAuthFlowDiff{}
}

func getOAuthFlowDiff(config *Config, flow1, flow2 *openapi3.OAuthFlow) (*OAuthFlowDiff, error) {
	diff, err := getOAuthFlowDiffInternal(config, flow1, flow2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getOAuthFlowDiffInternal(config *Config, flow1, flow2 *openapi3.OAuthFlow) (*OAuthFlowDiff, error) {

	if flow1 == nil && flow2 == nil {
		return nil, nil
	}

	result := OAuthFlowDiff{}
	var err error

	if flow1 == nil && flow2 != nil {
		result.Added = true
		return &result, nil
	}

	if flow1 != nil && flow2 == nil {
		result.Deleted = true
		return &result, nil
	}

	result.ExtensionsDiff, err = getExtensionsDiff(config, flow1.Extensions, flow2.Extensions)
	if err != nil {
		return nil, err
	}

	result.AuthorizationURLDiff = getValueDiff(flow1.AuthorizationURL, flow2.AuthorizationURL)
	result.TokenURLDiff = getValueDiff(flow1.TokenURL, flow2.TokenURL)
	result.RefreshURLDiff = getValueDiff(flow1.RefreshURL, flow2.RefreshURL)
	result.ScopesDiff = getStringMapDiff(flow1.Scopes, flow2.Scopes)

	return &result, nil
}
