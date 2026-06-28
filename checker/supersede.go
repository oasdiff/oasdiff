package checker

// PROTOTYPE (#702 follow-up): a declarative post-checker suppression pass.
//
// Several checks today suppress findings inline, each re-deriving from the diff
// that a higher-level finding already covers the same change. This pass moves
// the "headline finding supersedes its mechanical artifacts" kind of
// suppression out of the individual checks into one place, keyed on the typed
// ChangeLocation rather than on positional message Args.
//
// Scope of the prototype: only the oneOf-wrapping supersessions, matched at body
// level. ListOfTypes is the natural next entry; property-level supersession
// (and the moved-property suppression in the property-updated checks, which
// needs per-property identity from the diff) deliberately stay where they are.

// bodySupersedes maps a headline finding to the lower-level findings it makes
// redundant when both occur on the same body.
var bodySupersedes = map[string][]string{
	RequestBodyWrappedInOneOfId:  {RequestBodyOneOfAddedId, RequestBodyTypeGeneralizedId, RequestBodyTypeChangedId},
	ResponseBodyWrappedInOneOfId: {ResponseBodyOneOfAddedId, ResponseBodyTypeChangedId},
}

// bodyKey is the co-location key: the body a finding sits on, independent of any
// property path within it.
type bodyKey struct {
	direction      string
	path           string
	method         string
	mediaType      string
	responseStatus string
}

func bodyKeyOf(c ApiChange) (bodyKey, bool) {
	if c.Location == nil {
		return bodyKey{}, false
	}
	return bodyKey{
		direction:      c.Location.Direction,
		path:           c.Path,
		method:         c.Operation,
		mediaType:      c.Location.MediaType,
		responseStatus: c.Location.ResponseStatus,
	}, true
}

// suppressSuperseded drops findings made redundant by a co-located headline
// finding. Co-location is a typed-field compare, no Args parsing.
func suppressSuperseded(changes Changes) Changes {
	// For each superseded id, the set of body keys where a superseder is present.
	suppressedAt := map[string]map[bodyKey]bool{}
	for _, c := range changes {
		ac, ok := c.(ApiChange)
		if !ok {
			continue
		}
		superseded, found := bodySupersedes[ac.Id]
		if !found {
			continue
		}
		key, ok := bodyKeyOf(ac)
		if !ok {
			continue
		}
		for _, sid := range superseded {
			if suppressedAt[sid] == nil {
				suppressedAt[sid] = map[bodyKey]bool{}
			}
			suppressedAt[sid][key] = true
		}
	}

	if len(suppressedAt) == 0 {
		return changes
	}

	out := make(Changes, 0, len(changes))
	for _, c := range changes {
		if ac, ok := c.(ApiChange); ok {
			if keys, found := suppressedAt[ac.Id]; found {
				if key, ok := bodyKeyOf(ac); ok && keys[key] {
					continue
				}
			}
		}
		out = append(out, c)
	}
	return out
}
