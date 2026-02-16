package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// WebhooksDiff describes the changes between a pair of Webhooks objects (OpenAPI 3.1)
type WebhooksDiff struct {
	Added    []string                      `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string                      `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedWebhooks              `json:"modified,omitempty" yaml:"modified,omitempty"`
	Base     map[string]*openapi3.PathItem `json:"-" yaml:"-"`
	Revision map[string]*openapi3.PathItem `json:"-" yaml:"-"`
}

// ModifiedWebhooks is a map of webhook names to their respective diffs
type ModifiedWebhooks map[string]*PathDiff

// Empty indicates whether a change was found in this element
func (webhooksDiff *WebhooksDiff) Empty() bool {
	if webhooksDiff == nil {
		return true
	}

	return len(webhooksDiff.Added) == 0 &&
		len(webhooksDiff.Deleted) == 0 &&
		len(webhooksDiff.Modified) == 0
}

func newWebhooksDiff() *WebhooksDiff {
	return &WebhooksDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedWebhooks{},
	}
}

func getWebhooksDiff(config *Config, state *state, webhooks1, webhooks2 map[string]*openapi3.PathItem) (*WebhooksDiff, error) {
	// Return nil if both webhook maps are nil or empty
	if len(webhooks1) == 0 && len(webhooks2) == 0 {
		return nil, nil
	}

	diff, err := getWebhooksDiffInternal(config, state, webhooks1, webhooks2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getWebhooksDiffInternal(config *Config, state *state, webhooks1, webhooks2 map[string]*openapi3.PathItem) (*WebhooksDiff, error) {
	result := newWebhooksDiff()

	for name := range webhooks1 {
		if _, ok := webhooks2[name]; !ok {
			result.Deleted = append(result.Deleted, name)
		}
	}

	for name := range webhooks2 {
		if _, ok := webhooks1[name]; !ok {
			result.Added = append(result.Added, name)
		}
	}

	for name, pathItem1 := range webhooks1 {
		if pathItem2, ok := webhooks2[name]; ok {
			// Webhooks don't have path parameters, so we pass an empty PathParamsMap
			pathItemPair := &pathItemPair{
				PathItem1:     pathItem1,
				PathItem2:     pathItem2,
				PathParamsMap: PathParamsMap{},
			}
			pathDiff, err := getPathDiff(config, state, pathItemPair)
			if err != nil {
				return nil, err
			}
			if !pathDiff.Empty() {
				result.Modified[name] = pathDiff
			}
		}
	}

	result.Base = webhooks1
	result.Revision = webhooks2

	return result, nil
}

func (webhooksDiff *WebhooksDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(webhooksDiff.Added),
		Deleted:  len(webhooksDiff.Deleted),
		Modified: len(webhooksDiff.Modified),
	}
}
