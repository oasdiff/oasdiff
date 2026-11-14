package checker

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APIVersionChangedId = "api-version-changed"
)

// APIVersionUpdatedCheck reports a change to info.version. It says nothing
// about whether the change was the right one: that is the versioning policy's
// question (see versioning_policy.go), which the exit code can enforce with
// --fail-on MISSING-VERSION-BUMP.
func APIVersionUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	if diffReport.InfoDiff == nil || diffReport.InfoDiff.VersionDiff == nil {
		return result
	}

	return append(result, InfoChange{
		Id:    APIVersionChangedId,
		Level: config.getLogLevel(APIVersionChangedId),
		Args:  []any{fmt.Sprintf("%v", diffReport.InfoDiff.VersionDiff.From), fmt.Sprintf("%v", diffReport.InfoDiff.VersionDiff.To)},
	}.WithSources(
		infoVersionSource(diffReport.BaseInfo),
		infoVersionSource(diffReport.RevisionInfo),
	))
}

// infoVersionSource is the location of the version field in one spec's info
// object; nil when the spec was loaded without origin tracking.
func infoVersionSource(info *openapi3.Info) *Source {
	if info == nil {
		return nil
	}
	return location.SourceFromField(info.Origin, "version")
}
