package checker

import (
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
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

	// schema is the node the change was computed from, kept so that a
	// recognized transition there can claim the change (see
	// transition_claims.go).
	schema *diff.SchemaDiff
	// root is the payload or parameter schema the walk started from, and
	// propertyPath the path from it the change is reported at, kept so that a
	// change below a schema several properties reach can list the others (see
	// attachSharedSchemas).
	root         *diff.SchemaDiff
	propertyPath string

	// sharedSchema is rendered after Details rather than stored there, so a
	// check that sets its own details does not drop it.
	sharedSchema *SharedSchema

	// guards holds the document states observed at the change's location
	// (a readOnly or writeOnly property). capByGuards derives the level
	// from them and keeps only the ones that changed it, which GetComment
	// then explains.
	guards []Guard
}

// NewApiChange creates a new ApiChange
func NewApiChange(id string, config *Config, args []any, comment string, operationsSources diff.OperationsSourcesMap, operation *openapi3.Operation, method, path string) ApiChange {
	return ApiChange{
		Id:          id,
		Level:       config.getLogLevel(id),
		Args:        args,
		Comment:     comment,
		OperationId: operation.OperationID,
		Operation:   method,
		Path:        path,
		CommonChange: CommonChange{
			Attributes: getAttributes(config, operation),
		},
	}
}

// WithSchema returns a copy of the ApiChange that records the schema diff the
// change was found in, the root schema the walk started from, and the property
// path of the change (empty for a change to the root itself). A later call
// replaces all three.
func (a ApiChange) WithSchema(root *diff.SchemaDiff, schemaDiff *diff.SchemaDiff, propertyPath string) ApiChange {
	a.root, a.schema, a.propertyPath = root, schemaDiff, propertyPath
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

// withoutGuard removes a guard observed at the change's location, so the
// guard neither lowers the change's level nor adds its comment.
func (c ApiChange) withoutGuard(guard Guard) ApiChange {
	var guards []Guard
	for _, g := range c.guards {
		if g != guard {
			guards = append(guards, g)
		}
	}
	c.guards = guards
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

// WithAllOfDisclaimers adds the allOf disclaimers for this change, taking the
// change ID from c so it is written once at the call site.
func (c ApiChange) WithAllOfDisclaimers(underAllOf bool, schemaDiff *diff.SchemaDiff) ApiChange {
	return c.WithDisclaimers(allOfDisclaimers(underAllOf, schemaDiff, c.Id))
}

// WithDetails returns a copy of the ApiChange with Details set
func (c ApiChange) WithDetails(details string) ApiChange {
	c.Details = details
	return c
}

// WithSharedSchema returns a copy of the ApiChange in a schema several
// properties of its payload reach.
func (c ApiChange) WithSharedSchema(shared *SharedSchema) ApiChange {
	c.sharedSchema = shared
	return c
}

// GetSharedSchema is nil unless several of the payload's properties reach the
// schema the change is in or below.
func (c ApiChange) GetSharedSchema() *SharedSchema {
	return c.sharedSchema
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

func (c ApiChange) GetArgs() []any {
	return c.Args
}

// Text is the change's message, with each argument set apart by format.
func (c ApiChange) Text(l Localizer, format ArgFormat) string {
	return l(c.Id, formattedValues(c.Args, format)...) + c.getDetailsSuffix(l, format)
}

func (c ApiChange) GetUncolorizedText(l Localizer) string {
	return c.Text(l, Quoted)
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

// getDetailsSuffix takes the format applied to the message arguments, so a
// value in the details looks the same as one in the message.
func (c ApiChange) getDetailsSuffix(l Localizer, format ArgFormat) string {
	details := combineDetails(c.Details, c.sharedSchema.detail(l, format))
	if details == "" {
		return ""
	}
	return " " + details
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

// Subject is the API the change was found on.
func (c ApiChange) Subject() (section, element string) {
	return "API", c.Operation + " " + c.Path
}

func (c ApiChange) GetPath() string {
	return c.Path
}
