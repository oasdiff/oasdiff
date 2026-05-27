package formatters

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
)

type JSONFormatter struct {
	notImplementedFormatter
	Localizer checker.Localizer
}

func newJSONFormatter(l checker.Localizer) JSONFormatter {
	return JSONFormatter{
		Localizer: l,
	}
}

func (f JSONFormatter) RenderDiff(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	return printJSON(diff)
}

func (f JSONFormatter) RenderSummary(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	return printJSON(diff.GetSummary())
}

func (f JSONFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, _, _ string) ([]byte, error) {
	return printJSON(adaptStructure(NewChanges(changes, f.Localizer), opts))
}

func (f JSONFormatter) RenderChecks(checks Checks, opts RenderOpts) ([]byte, error) {
	return printJSON(checks)
}

func (f JSONFormatter) RenderFlatten(spec *openapi3.T, opts RenderOpts) ([]byte, error) {
	return printJSON(spec)
}

func (f JSONFormatter) RenderValidate(findings Findings, opts RenderOpts) ([]byte, error) {
	return printJSON(findings)
}

func (f JSONFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputSummary, OutputChangelog, OutputChecks, OutputFlatten, OutputValidate}
}

func printJSON(output any) ([]byte, error) {
	if reflect.ValueOf(output).IsNil() {
		return nil, nil
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return bytes, nil
}

// adaptStructure wraps the changes list in an object when the caller asks
// for the wrapped shape (used by oasdiff-service so the response carries
// extra signals alongside the change list). The bare-array shape ignores
// opts to preserve the existing CLI JSON output.
//
// diff_empty is true when the underlying diff found no changes at all,
// so consumers can distinguish "specs are identical" from "specs differ
// but no breaking-change / changelog rule fired."
func adaptStructure(output any, opts RenderOpts) any {
	if opts.WrapInObject {
		return map[string]any{
			"changes":    output,
			"diff_empty": opts.DiffEmpty,
		}
	}
	return output
}
