package internal

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/invopop/jsonschema"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

const schemaCmd = "schema"

func getSchemaCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "schema [flags]",
		Short: "Display the JSON Schema of the JSON output",
		Long: `Print a JSON Schema for the "--format json" output of the breaking and changelog commands.
The "--format yaml" output has the same structure.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions, // see https://github.com/spf13/cobra/issues/1969
		RunE:              getRun(runSchema),
	}

	return &cmd
}

func runSchema(_ *Flags, stdout io.Writer) (bool, *ReturnError) {
	return false, outputSchema(stdout)
}

func outputSchema(stdout io.Writer) *ReturnError {
	out, err := json.MarshalIndent(changesJSONSchema(), "", "  ")
	if err != nil {
		return getErrFailedPrint(schemaCmd, err)
	}
	if _, err := fmt.Fprintln(stdout, string(out)); err != nil {
		return getErrFailedPrint(schemaCmd, err)
	}
	return nil
}

// changesJSONSchema reflects the JSON Schema of the breaking/changelog output
// (formatters.Changes) directly from its Go type, so the schema and the actual
// "--format json" output never drift apart.
func changesJSONSchema() *jsonschema.Schema {
	return (&jsonschema.Reflector{}).Reflect(formatters.Changes{})
}
