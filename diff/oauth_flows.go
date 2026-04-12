package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// OAuthFlowsDiff describes the changes between a pair of oauth flows objects: https://swagger.io/specification/#oauth-flows-object
type OAuthFlowsDiff struct {
	Added                 bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted               bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff        *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	ImplicitDiff          *OAuthFlowDiff  `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	PasswordDiff          *OAuthFlowDiff  `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentialsDiff *OAuthFlowDiff  `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCodeDiff *OAuthFlowDiff  `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *OAuthFlowsDiff) Empty() bool {
	return diff == nil || *diff == OAuthFlowsDiff{}
}

func getOAuthFlowsDiff(config *Config, flows1, flows2 *openapi3.OAuthFlows) (*OAuthFlowsDiff, error) {
	diff, err := getOAuthFlowsDiffInternal(config, flows1, flows2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getOAuthFlowsDiffInternal(config *Config, flows1, flows2 *openapi3.OAuthFlows) (*OAuthFlowsDiff, error) {

	if flows1 == nil && flows2 == nil {
		return nil, nil
	}

	if flows1 == nil && flows2 != nil {
		return &OAuthFlowsDiff{
			Added: true,
		}, nil
	}

	if flows1 != nil && flows2 == nil {
		return &OAuthFlowsDiff{
			Deleted: true,
		}, nil
	}

	result := OAuthFlowsDiff{}
	var err error

	result.ExtensionsDiff, err = getExtensionsDiff(config, flows1.Extensions, flows2.Extensions)
	if err != nil {
		return nil, err
	}

	result.ImplicitDiff, err = getOAuthFlowDiff(config, flows1.Implicit, flows2.Implicit)
	if err != nil {
		return nil, err
	}

	result.PasswordDiff, err = getOAuthFlowDiff(config, flows1.Password, flows2.Password)
	if err != nil {
		return nil, err
	}

	result.ClientCredentialsDiff, err = getOAuthFlowDiff(config, flows1.ClientCredentials, flows2.ClientCredentials)
	if err != nil {
		return nil, err
	}

	result.AuthorizationCodeDiff, err = getOAuthFlowDiff(config, flows1.AuthorizationCode, flows2.AuthorizationCode)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
