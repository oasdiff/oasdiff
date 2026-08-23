package coverage

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/stretchr/testify/require"
)

// suggestId names the check an unchecked edit would need: a context prefix
// from the location, the edited keyword, and the action as a past-tense
// verb.
func TestSuggestId(t *testing.T) {
	for _, tc := range []struct {
		location string
		action   metaschema.Action
		want     string
	}{
		// a field names itself, in kebab case
		{"paths.*.*.requestBody.content.*.schema.maxLength", metaschema.ActionDecrease, "request-body-max-length-decreased"},
		{"paths.*.*.parameters.*.explode", metaschema.ActionSet, "request-parameter-explode-set"},
		{"paths.*.*.responses.*.headers.*.schema.pattern", metaschema.ActionChange, "response-header-pattern-changed"},
		{"components.securitySchemes.*.openIdConnectUrl", metaschema.ActionChange, "api-security-scheme-open-id-connect-url-changed"},
		{"components.schemas.*.multipleOf", metaschema.ActionUnset, "api-component-multiple-of-unset"},
		{"webhooks.*.*.requestBody", metaschema.ActionSet, "webhook-request-body-set"},

		// a map entry takes its collection's name, singular
		{"paths.*", metaschema.ActionRemove, "api-path-removed"},
		{"paths.*.*.parameters.*", metaschema.ActionAdd, "request-parameter-added"},
		{"paths.*.*.responses.*", metaschema.ActionRemove, "response-removed"},

		// nested entries walk back to the nearest named ancestor rather than
		// naming themselves after a wildcard
		{"webhooks.*.*", metaschema.ActionAdd, "webhook-added"},
		{"paths.*.*.callbacks.*.*", metaschema.ActionAdd, "callback-added"},

		// a prefix that already names the ancestor is not repeated
		{"paths.*.*.callbacks.*", metaschema.ActionAdd, "callback-added"},
		{"paths.*.*.requestBody", metaschema.ActionSet, "request-body-set"},

		// a $-prefixed keyword drops the sigil
		{"paths.*.*.requestBody.content.*.schema.$defs.*", metaschema.ActionAdd, "request-body-def-added"},
	} {
		require.Equal(t, tc.want, suggestId(metaschema.Edit{Location: tc.location, Action: tc.action}), tc.location)
	}
}
