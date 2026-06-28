package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// PROTOTYPE: the post-pass keys on typed ChangeLocation, so it suppresses
// superseded findings only on the same body and leaves a different body alone.
func TestSuppressSuperseded_prototype(t *testing.T) {
	body := func(id, dir, status string) ApiChange {
		return ApiChange{
			Id:        id,
			Path:      "/x",
			Operation: "GET",
			Location:  &ChangeLocation{Direction: dir, MediaType: "application/json", ResponseStatus: status},
		}
	}

	in := Changes{
		body(ResponseBodyWrappedInOneOfId, "response", "200"), // headline
		body(ResponseBodyOneOfAddedId, "response", "200"),     // co-located -> dropped
		body(ResponseBodyTypeChangedId, "response", "200"),    // co-located -> dropped
		body(ResponseBodyOneOfAddedId, "response", "404"),     // different body -> kept
	}

	out := suppressSuperseded(in)

	ids := make([]string, 0, len(out))
	for _, c := range out {
		ids = append(ids, c.GetId())
	}
	require.ElementsMatch(t, []string{
		ResponseBodyWrappedInOneOfId,
		ResponseBodyOneOfAddedId, // the 404 one
	}, ids)
}

// Without a headline finding present, nothing is suppressed.
func TestSuppressSuperseded_noHeadline_prototype(t *testing.T) {
	in := Changes{
		ApiChange{Id: ResponseBodyOneOfAddedId, Path: "/x", Operation: "GET",
			Location: &ChangeLocation{Direction: "response", ResponseStatus: "200"}},
	}
	require.Len(t, suppressSuperseded(in), 1)
}
