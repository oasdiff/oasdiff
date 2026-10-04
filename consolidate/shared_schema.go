package consolidate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
)

// SharedSchema merges the findings that one check reports, in one operation
// and payload, for one change in a schema several properties of the payload
// reference. The checker reports the change at each reference. The finding
// at the property first in alphabetical order is kept, in the place of the
// first of its group; it already lists the other properties, and a comment
// says why they are not reported separately.
func SharedSchema(changes checker.Changes) checker.Changes {
	kept := map[string]int{}
	result := make(checker.Changes, 0, len(changes))
	for _, change := range changes {
		apiChange, ok := change.(checker.ApiChange)
		if !ok || apiChange.GetSharedSchema() == nil {
			result = append(result, change)
			continue
		}
		if apiChange.Comment == "" {
			apiChange.Comment = checker.SharedSchemaCommentId
		}
		key := sharedSchemaKey(apiChange)
		i, ok := kept[key]
		if !ok {
			kept[key] = len(result)
			result = append(result, apiChange)
			continue
		}
		if apiChange.GetSharedSchema().Properties[0] < result[i].(checker.ApiChange).GetSharedSchema().Properties[0] {
			result[i] = apiChange
		}
	}
	return result
}

// sharedSchemaKey is the same for the findings of one change at different
// references: they differ only in the property each is reported at, which
// is left out, and the shared schema lists the same properties for all.
func sharedSchemaKey(change checker.ApiChange) string {
	shared := change.GetSharedSchema()
	args := make([]string, len(change.Args))
	for i, arg := range change.Args {
		if s := fmt.Sprintf("%v", arg); s != shared.Properties[0] {
			args[i] = s
		}
	}
	properties := slices.Sorted(slices.Values(shared.Properties))
	return strings.Join([]string{
		change.Operation, change.Path, change.Id, change.Details, shared.Name,
		strings.Join(properties, "\x00"), strings.Join(args, "\x00"),
	}, "\x01")
}
