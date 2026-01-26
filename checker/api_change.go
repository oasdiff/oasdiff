package checker

import (
	"fmt"
	"strings"

	"github.com/TwiN/go-color"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

// ApiChange represnts a change in the Paths Section of an OpenAPI spec
type ApiChange struct {
	CommonChange

	Id          string
	Args        []any
	Comment     string
	Details     string
	Level       Level
	Operation   string
	OperationId string
	Path        string
	Source      *load.Source

	SourceFile      string
	SourceLine      int
	SourceLineEnd   int
	SourceColumn    int
	SourceColumnEnd int
}

// NewApiChange creates a new ApiChange
// TODO: use opInfo to simplify the function signature
func NewApiChange(id string, config *Config, args []any, comment string, operationsSources *diff.OperationsSourcesMap, operation *openapi3.Operation, method, path string) ApiChange {
	return ApiChange{
		Id:          id,
		Level:       config.getLogLevel(id),
		Args:        args,
		Comment:     comment,
		OperationId: operation.OperationID,
		Operation:   method,
		Path:        path,
		Source:      load.NewSource((*operationsSources)[operation]),
		CommonChange: CommonChange{
			Attributes: getAttributes(config, operation),
		},
	}
}

// WithDetails returns a copy of the ApiChange with Details set
func (c ApiChange) WithDetails(details string) ApiChange {
	c.Details = details
	return c
}

func getAttributes(config *Config, operation *openapi3.Operation) map[string]any {
	result := map[string]any{}
	for _, tag := range config.Attributes {
		if val, ok := operation.Extensions[tag]; ok {
			result[tag] = val
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func (c ApiChange) GetSection() string {
	return "paths"
}

func (c ApiChange) IsBreaking() bool {
	return c.GetLevel().IsBreaking()
}

func (c ApiChange) MatchIgnore(ignorePath, ignoreLine string, l Localizer) bool {
	if ignorePath == "" {
		return false
	}

	return ignorePath == strings.ToLower(c.Path) &&
		strings.Contains(ignoreLine, strings.ToLower(c.Operation+" "+c.Path)) &&
		strings.Contains(ignoreLine, strings.ToLower(c.GetUncolorizedText(l)))
}

func (c ApiChange) GetId() string {
	return c.Id
}

func (c ApiChange) GetText(l Localizer) string {
	return l(c.Id, colorizedValues(c.Args)...) + c.getDetailsSuffix()
}

func (c ApiChange) GetArgs() []any {
	return c.Args
}

func (c ApiChange) GetUncolorizedText(l Localizer) string {
	return l(c.Id, quotedValues(c.Args)...) + c.getDetailsSuffix()
}

func (c ApiChange) GetComment(l Localizer) string {
	return l(c.Comment)
}

func (c ApiChange) GetDetails() string {
	return c.Details
}

func (c ApiChange) getDetailsSuffix() string {
	if c.Details == "" {
		return ""
	}
	return " " + c.Details
}

func (c ApiChange) GetLevel() Level {
	return c.Level
}

func (c ApiChange) GetOperation() string {
	return c.Operation
}

func (c ApiChange) GetOperationId() string {
	return c.OperationId
}

func (c ApiChange) GetPath() string {
	return c.Path
}

func (c ApiChange) GetSource() string {
	return c.Source.String()
}

func (c ApiChange) GetSourceFile() string {
	if c.SourceFile != "" {
		return c.SourceFile
	}

	if c.Source.IsFile() {
		return c.Source.String()
	}

	return ""
}

func (c ApiChange) GetSourceLine() int {
	return c.SourceLine
}

func (c ApiChange) GetSourceLineEnd() int {
	return c.SourceLineEnd
}

func (c ApiChange) GetSourceColumn() int {
	return c.SourceColumn
}

func (c ApiChange) GetSourceColumnEnd() int {
	return c.SourceColumnEnd
}

func (c ApiChange) SingleLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s %s %s, %s API %s %s %s [%s]. %s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), l("at"), c.GetSource(), l("in"), color.InGreen(c.Operation), color.InGreen(c.Path), c.GetText(l), color.InYellow(c.Id), c.GetComment(l))
	}

	return fmt.Sprintf(format, c.Level.String(), l("at"), c.GetSource(), l("in"), c.Operation, c.Path, c.GetUncolorizedText(l), c.Id, c.GetComment(l))

}

func (c ApiChange) MultiLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s\t[%s] %s %s\t\n\t%s API %s %s\n\t\t%s%s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), color.InYellow(c.Id), l("at"), c.GetSource(), l("in"), color.InGreen(c.Operation), color.InGreen(c.Path), c.GetText(l), multiLineComment(c.GetComment(l)))
	}

	return fmt.Sprintf(format, c.Level.String(), c.Id, l("at"), c.GetSource(), l("in"), c.Operation, c.Path, c.GetUncolorizedText(l), multiLineComment(c.GetComment(l)))
}

func multiLineComment(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("\n\t\t%s", comment)
}
