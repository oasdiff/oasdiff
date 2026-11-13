package validate

import (
	"fmt"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// AllOfConflictID reports subschemas under allOf that cannot be merged into
// one schema. Merging is what lets the rules see the schema a payload is
// actually validated against, so when it fails that schema goes unchecked.
//
// Either the subschemas contradict each other, in which case no payload
// satisfies them all, or the merge has no rule for the combination. The
// finding reports the conflict without claiming which, because the merge
// cannot tell them apart.
//
// Only a caller that merges allOf before validating can report it, so it
// never fires on `oasdiff validate` without --flatten-allof.
const AllOfConflictID = "allof-subschemas-conflict"

// AllOfConflict returns the finding for a failed merge, so a caller that
// validates a flattened spec can report the failure and carry on with the
// spec as written. reason is the error the merge returned, and source is the
// spec's display name, as in Validate.
//
// The finding carries no line or column: the merge reports which conflict it
// hit, not where.
func AllOfConflict(reason error, source string) formatters.Finding {
	finding := formatters.Finding{
		Id:    AllOfConflictID,
		Text:  fmt.Sprintf("subschemas under allOf cannot be merged, so the merged schema was not validated: %s", reason),
		Level: checker.WARN,
		Source: formatters.Source{
			File: source,
		},
	}
	finding.Fingerprint = checker.ComputeFingerprint(finding.Id, "", "", []any{reason.Error()})
	return finding
}
