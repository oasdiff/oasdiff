package coverage

import (
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/utils"
)

// suggestId derives a candidate check id for an unchecked edit, in the
// naming style of the existing ids: a context prefix from the location, the
// edited keyword, and the action as a past-tense verb. It is a hint for
// naming the missing check, in the spirit of the retired generator (#1168).
func suggestId(edit metaschema.Edit) string {
	verb := map[metaschema.Action]string{
		metaschema.ActionAdd:      "added",
		metaschema.ActionRemove:   "removed",
		metaschema.ActionChange:   "changed",
		metaschema.ActionSet:      "set",
		metaschema.ActionUnset:    "unset",
		metaschema.ActionIncrease: "increased",
		metaschema.ActionDecrease: "decreased",
	}[edit.Action]

	loc := edit.Location
	var prefix string
	switch {
	case strings.HasPrefix(loc, "components.securitySchemes"):
		prefix = "api-security-scheme"
	case strings.HasPrefix(loc, "components"):
		prefix = "api-component"
	case strings.HasPrefix(loc, "webhooks"):
		prefix = "webhook"
	case strings.HasPrefix(loc, "security"):
		prefix = "api-security"
	case strings.Contains(loc, ".callbacks."):
		prefix = "callback"
	case strings.Contains(loc, ".responses.") && strings.Contains(loc, ".headers."):
		prefix = "response-header"
	case strings.Contains(loc, ".responses."):
		prefix = "response"
	case strings.Contains(loc, ".requestBody"):
		prefix = "request-body"
	case strings.Contains(loc, ".parameters."):
		prefix = "request-parameter"
	case strings.HasPrefix(loc, "paths."):
		prefix = "api"
	default:
		prefix = strings.SplitN(loc, ".", 2)[0]
	}

	// the edited keyword: the last segment that names a field. A map entry
	// is a wildcard, so it takes the name of its nearest named ancestor,
	// singular; nested entries under one collection therefore share a hint.
	segs := strings.Split(loc, ".")
	var keyword string
	for i, seg := range slices.Backward(segs) {
		if seg == "*" || seg == "x-*" {
			continue
		}
		keyword = strings.TrimPrefix(seg, "$")
		if i < len(segs)-1 {
			keyword = strings.TrimSuffix(keyword, "s")
		}
		break
	}

	// the nearest named ancestor can be the one the prefix already names
	kebab := utils.ToKebabCase(keyword)
	if prefix == kebab || strings.HasSuffix(prefix, "-"+kebab) {
		return prefix + "-" + verb
	}
	return prefix + "-" + kebab + "-" + verb
}
