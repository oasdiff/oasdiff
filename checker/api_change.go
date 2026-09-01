package checker

import (
	"fmt"
	"slices"
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
	Disclaimers []Disclaimer
	Details     string
	Level       Level
	Operation   string
	OperationId string
	Path        string
	Source      *load.Source

	// claimed marks a change that a recognized schema transition explains
	// (see transition_claims.go); claimed changes are dropped in favor of the
	// transition's own finding. Set by WithSchema.
	claimed bool

	// guards holds the document states observed at the change's location
	// (a readOnly or writeOnly property). capByGuards derives the level
	// from them and keeps only the ones that changed it, which GetComment
	// then explains.
	guards []Guard

	// DEPRECATED: Will be removed after migration to BaseSource/RevisionSource
	SourceFile      string
	SourceLine      int
	SourceLineEnd   int
	SourceColumn    int
	SourceColumnEnd int
}

// NewApiChange creates a new ApiChange
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

// WithSchema returns a copy of the ApiChange with claimed set: the change is
// claimed when a recognized transition at the given schema node (the node the
// change was computed from) claims the change's rule (see
// transition_claims.go). The node itself is not retained.
func (a ApiChange) WithSchema(schemaDiff *diff.SchemaDiff) ApiChange {
	a.claimed = claimedByTransition(schemaDiff, a.Id)
	return a
}

// WithSources returns a copy of the ApiChange with BaseSource and RevisionSource populated
func (a ApiChange) WithSources(baseSource, revisionSource *Source) ApiChange {
	a.BaseSource = baseSource
	a.RevisionSource = revisionSource
	return a
}

// WithGuards is additive and drops duplicates, like WithDisclaimers.
func (c ApiChange) WithGuards(guards []Guard) ApiChange {
	for _, g := range guards {
		if !slices.Contains(c.guards, g) {
			c.guards = append(c.guards, g)
		}
	}
	return c
}

// WithDisclaimers is additive and drops duplicates: a change can gather the
// same condition from more than one place.
func (c ApiChange) WithDisclaimers(disclaimers []Disclaimer) ApiChange {
	for _, d := range disclaimers {
		if !slices.Contains(c.Disclaimers, d) {
			c.Disclaimers = append(c.Disclaimers, d)
		}
	}
	return c
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
	parts := make([]string, 0, 1+len(c.Disclaimers))
	if c.Comment != "" {
		parts = append(parts, l(c.Comment))
	}
	for _, d := range c.Disclaimers {
		parts = append(parts, l(commentId(d.String())))
	}
	for _, g := range c.guards {
		parts = append(parts, l(commentId(string(g))))
	}
	return strings.Join(parts, " ")
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

func (c ApiChange) GetDisclaimers() []Disclaimer {
	return c.Disclaimers
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
	return c.Source.DisplayPath()
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
	const format = "%s\t[%s] %s %s\n\t%s API %s %s\n\t\t%s%s"

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
