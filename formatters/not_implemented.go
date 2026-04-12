package formatters

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

type notImplementedFormatter struct{}

func (f notImplementedFormatter) RenderDiff(*diff.Diff, RenderOpts) ([]byte, error) {
	return notImplemented()
}

func (f notImplementedFormatter) RenderSummary(*diff.Diff, RenderOpts) ([]byte, error) {
	return notImplemented()
}

func (f notImplementedFormatter) RenderChangelog(checker.Changes, RenderOpts, *load.SpecInfoPair) ([]byte, error) {
	return notImplemented()
}

func (f notImplementedFormatter) RenderChecks(Checks, RenderOpts) ([]byte, error) {
	return notImplemented()
}

func (f notImplementedFormatter) RenderFlatten(*openapi3.T, RenderOpts) ([]byte, error) {
	return notImplemented()
}

func (f notImplementedFormatter) SupportsTemplate() bool {
	return false
}

func notImplemented() ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
