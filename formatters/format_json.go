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
	return printJSON(adaptStructure(NewChanges(changes, f.Localizer), opts.WrapInObject))
}

func (f JSONFormatter) RenderChecks(checks Checks, opts RenderOpts) ([]byte, error) {
	return printJSON(checks)
}

func (f JSONFormatter) RenderFlatten(spec *openapi3.T, opts RenderOpts) ([]byte, error) {
	return printJSON(spec)
}

func (f JSONFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputSummary, OutputChangelog, OutputChecks, OutputFlatten}
}

func printJSON(output interface{}) ([]byte, error) {
	if reflect.ValueOf(output).IsNil() {
		return nil, nil
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return bytes, nil
}

func adaptStructure(output any, wrapInObject bool) any {
	if wrapInObject {
		output = map[string]any{"changes": output}
	}

	return output
}
