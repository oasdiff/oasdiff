package formatters

import (
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"gopkg.in/yaml.v3"
)

type YAMLFormatter struct {
	Localizer checker.Localizer
}

func newYAMLFormatter(l checker.Localizer) YAMLFormatter {
	return YAMLFormatter{
		Localizer: l,
	}
}

func (f YAMLFormatter) RenderDiff(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	return printYAML(diff)
}

func (f YAMLFormatter) RenderSummary(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	return printYAML(diff.GetSummary())
}

func (f YAMLFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, specInfoPair *load.SpecInfoPair) ([]byte, error) {
	return printYAML(adaptStructure(NewChanges(changes, f.Localizer), opts.WrapInObject))
}

func (f YAMLFormatter) RenderChecks(checks Checks, opts RenderOpts) ([]byte, error) {
	return printYAML(checks)
}

func (f YAMLFormatter) RenderFlatten(spec *openapi3.T, opts RenderOpts) ([]byte, error) {
	return printYAML(spec)
}

func (f YAMLFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputSummary, OutputChangelog, OutputChecks, OutputFlatten}
}

func printYAML(output interface{}) ([]byte, error) {
	if reflect.ValueOf(output).IsNil() {
		return nil, nil
	}

	bytes, err := yaml.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return bytes, nil
}
