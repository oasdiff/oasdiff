package checker_test

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/internal/populatetest"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// boundCell is one action cell of a bound keyword, read from a rule's
// public claims and id: which side, which schema root, which keyword, and
// the registered level.
type boundCell struct {
	id        string
	direction string // "request" | "response"
	scope     string // "body" | "property" | "parameter" | "header"
	keyword   string // the bound keyword, e.g. "maximum"
	action    string // "set" | "unset" | "increased" | "decreased"
	level     checker.Level
}

// boundCellVerbs maps a claim's action to the id verb of its cell. multipleOf
// increase and decrease claims map to no cell: numeric order does not decide
// their effect, so their rules (multiple-of-changed and friends) are not
// bound cells.
func boundCellVerbs(keyword, claimAction string) (string, bool) {
	if keyword == "multipleOf" && claimAction != "set" && claimAction != "unset" {
		return "", false
	}
	verb, ok := map[string]string{
		"set":      "set",
		"unset":    "unset",
		"increase": "increased",
		"decrease": "decreased",
	}[claimAction]
	return verb, ok
}

var boundClaimRe = regexp.MustCompile(`^paths\.\*\.\*\.(requestBody\.content\.\*|responses\.\*\.content\.\*|parameters\.\*|responses\.\*\.headers\.\*)\.schema\.(\w+):(.+)$`)

// boundCells enumerates every action cell of every bound keyword from
// public data alone: diff.SchemaBounds names the keywords and each rule's
// claims and id say which cells it covers. Hand-written and generated rules
// are indistinguishable here, deliberately: the gate holds for both. Guarded
// rules are skipped: they are variants of a cell fired only under their
// guard, and the fire docs carry no guard.
func boundCells(t *testing.T) []boundCell {
	t.Helper()

	keywords := map[string]bool{}
	for _, bound := range diff.SchemaBounds {
		keywords[bound.Keyword] = true
	}

	var cells []boundCell
	for _, rule := range checker.GetAllRules() {
		if len(rule.Guards) > 0 {
			continue
		}
		for _, loc := range rule.Locations {
			cells = append(cells, boundClaimCells(t, rule, loc, keywords)...)
		}
	}
	return cells
}

// boundClaimCells reads one claim as set/unset bound cells: the direction
// and schema root from the location (body and property share a location, so
// the rule's id disambiguates them), the keyword, and one cell per claimed
// set or unset action. A claim not about a bound keyword yields none.
func boundClaimCells(t *testing.T, rule checker.BackwardCompatibilityRule, claim string, keywords map[string]bool) []boundCell {
	t.Helper()

	m := boundClaimRe.FindStringSubmatch(claim)
	if m == nil || !keywords[m[2]] {
		return nil
	}

	var direction, scope string
	switch m[1] {
	case "parameters.*":
		direction, scope = "request", "parameter"
	case "responses.*.headers.*":
		direction, scope = "response", "header"
	case "requestBody.content.*":
		direction = "request"
	case "responses.*.content.*":
		direction = "response"
	}
	if scope == "" {
		switch {
		case strings.Contains(rule.Id, "-body-"):
			scope = "body"
		case strings.Contains(rule.Id, "-property-"):
			scope = "property"
		default:
			t.Fatalf("%s claims a bound edit but its id names neither body nor property", rule.Id)
		}
	}

	var cells []boundCell
	for action := range strings.SplitSeq(m[3], ",") {
		verb, ok := boundCellVerbs(m[2], action)
		if !ok {
			continue
		}
		cells = append(cells, boundCell{
			id:        rule.Id,
			direction: direction,
			scope:     scope,
			keyword:   m[2],
			action:    verb,
			level:     rule.Level,
		})
	}
	return cells
}

// setBoundSample populates the schema field whose json tag is the keyword,
// deriving the sample from the field's type; a larger magnitude yields a
// larger value.
func setBoundSample(t *testing.T, s *openapi3.Schema, keyword string, magnitude uint64) {
	t.Helper()
	typ := reflect.TypeFor[openapi3.Schema]()
	for field := range typ.Fields() {
		if name, _, _ := strings.Cut(field.Tag.Get("json"), ","); name == keyword {
			require.True(t, populatetest.NonZeroScale(reflect.ValueOf(s).Elem().FieldByName(field.Name), keyword, magnitude), keyword)
			return
		}
	}
	t.Fatalf("no openapi3.Schema field with json tag %q", keyword)
}

// boundCellDoc builds a spec that carries the keyword sample at the cell's
// schema root; magnitude 0 leaves the keyword absent.
func boundCellDoc(t *testing.T, cell boundCell, magnitude uint64) *load.SpecInfo {
	t.Helper()
	node := &openapi3.Schema{}
	if magnitude > 0 {
		setBoundSample(t, node, cell.keyword, magnitude)
	}
	carrier := node
	if cell.scope == "property" {
		carrier = &openapi3.Schema{
			Properties: openapi3.Schemas{"p": &openapi3.SchemaRef{Value: node}},
		}
	}

	plain := func() *openapi3.SchemaRef { return &openapi3.SchemaRef{Value: &openapi3.Schema{}} }
	requestSchema, responseSchema, parameterSchema, headerSchema := plain(), plain(), plain(), plain()
	switch cell.scope {
	case "parameter":
		parameterSchema = &openapi3.SchemaRef{Value: carrier}
	case "header":
		headerSchema = &openapi3.SchemaRef{Value: carrier}
	case "body", "property":
		if cell.direction == "response" {
			responseSchema = &openapi3.SchemaRef{Value: carrier}
		} else {
			requestSchema = &openapi3.SchemaRef{Value: carrier}
		}
	}

	op := &openapi3.Operation{
		Parameters: openapi3.Parameters{&openapi3.ParameterRef{Value: &openapi3.Parameter{
			Name: "p", In: "query", Schema: parameterSchema,
		}}},
		RequestBody: &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{
			Content: openapi3.Content{"application/json": &openapi3.MediaType{Schema: requestSchema}},
		}},
		Responses: openapi3.NewResponses(openapi3.WithStatus(200, &openapi3.ResponseRef{Value: &openapi3.Response{
			Description: new("ok"),
			Content:     openapi3.Content{"application/json": &openapi3.MediaType{Schema: responseSchema}},
			Headers: openapi3.Headers{"X-Rate-Limit": &openapi3.HeaderRef{Value: &openapi3.Header{
				Parameter: openapi3.Parameter{Schema: headerSchema},
			}}},
		}})),
	}

	return &load.SpecInfo{Spec: &openapi3.T{
		OpenAPI: "3.1.0",
		Info:    &openapi3.Info{Title: "t", Version: "1.0.0"},
		Paths:   openapi3.NewPaths(openapi3.WithPath("/t", &openapi3.PathItem{Post: op})),
	}}
}

// Every action cell of every bound keyword fires: for each cell read
// from the public rule claims, a spec pair built from the keyword's sample
// produces exactly one change, with the cell's id at its registered level
// and a message that renders rather than echoing its key. All checks run,
// so a second check covering the same cell under another id fails here.
func TestBoundCellsFire(t *testing.T) {
	cells := boundCells(t)
	require.Len(t, cells, 300)

	localizer := checker.NewDefaultLocalizer()
	config := allChecksConfig()

	magnitudes := map[string][2]uint64{
		"set":       {0, 1},
		"unset":     {1, 0},
		"increased": {1, 2},
		"decreased": {2, 1},
	}
	for _, cell := range cells {
		pair := magnitudes[cell.action]
		base := boundCellDoc(t, cell, pair[0])
		revision := boundCellDoc(t, cell, pair[1])

		d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
		require.NoError(t, err, cell.id)
		changes := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

		require.Len(t, changes, 1, "%s (%s %s): the edit must be reported exactly once", cell.id, cell.keyword, cell.action)
		change := changes[0]
		require.Equal(t, cell.id, change.GetId())
		require.Equal(t, cell.level, change.GetLevel(), cell.id)
		text := change.GetUncolorizedText(localizer)
		require.NotContains(t, text, cell.id, "message must render, not echo its key: %s", text)
		require.NotContains(t, change.GetComment(localizer), "-comment", cell.id)
	}
}

// The registered level of a sample of generated cells, stated as the
// contract judgment rather than re-derived: raising a lower bound or
// lowering an upper bound rejects request payloads the old contract
// accepted, and widening a response bound sends values old clients never
// had to handle. The fire test checks each change against its rule's level,
// so a wrong polarity row would stay self-consistent; these anchors break
// that symmetry.
func TestBoundLevelAnchors(t *testing.T) {
	expected := map[string]checker.Level{
		"request-parameter-min-properties-increased": checker.ERR,
		"request-body-min-items-decreased":           checker.INFO,
		"response-body-max-decreased":                checker.INFO,
		"response-property-min-increased":            checker.INFO,
		"response-header-max-increased":              checker.ERR,
		"response-header-max-length-decreased":       checker.INFO,
	}
	for _, rule := range checker.GetAllRules() {
		if level, ok := expected[rule.Id]; ok {
			require.Equal(t, level, rule.Level, rule.Id)
			delete(expected, rule.Id)
		}
	}
	require.Empty(t, expected, "anchored ids missing from GetAllRules")
}

// exclusiveBoolDoc builds a spec whose request body has a property with the
// OpenAPI 3.0 boolean form of exclusiveMaximum, or none.
func exclusiveBoolDoc(set *bool) *load.SpecInfo {
	prop := &openapi3.Schema{Type: &openapi3.Types{"integer"}}
	if set != nil {
		prop.ExclusiveMax = openapi3.ExclusiveBound{Bool: set}
	}
	carrier := &openapi3.Schema{Properties: openapi3.Schemas{"p": &openapi3.SchemaRef{Value: prop}}}
	op := &openapi3.Operation{
		RequestBody: &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{
			Content: openapi3.Content{"application/json": &openapi3.MediaType{Schema: &openapi3.SchemaRef{Value: carrier}}},
		}},
		Responses: openapi3.NewResponses(openapi3.WithStatus(200, &openapi3.ResponseRef{Value: &openapi3.Response{
			Description: new("ok"),
		}})),
	}
	return &load.SpecInfo{Spec: &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &openapi3.Info{Title: "t", Version: "1.0.0"},
		Paths:   openapi3.NewPaths(openapi3.WithPath("/t", &openapi3.PathItem{Post: op})),
	}}
}

func exclusiveBoolChanges(t *testing.T, base, revision *bool) checker.Changes {
	t.Helper()
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), exclusiveBoolDoc(base), exclusiveBoolDoc(revision))
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
}

// The OpenAPI 3.0 boolean form through the checks. exclusiveMaximum: false
// declares the bound not exclusive, the same contract as leaving it out, so
// false to true is the set it always was in effect, adding false reports
// nothing, and removing true is the generated unset. The first two pin the
// bugs the bound classification fixed: false to true reported nothing, and
// adding false reported a breaking set.
func TestBound_ExclusiveBooleanForm(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }

	falseToTrue := exclusiveBoolChanges(t, boolPtr(false), boolPtr(true))
	require.Len(t, falseToTrue, 1)
	require.Equal(t, checker.RequestPropertyExclusiveMaxSetId, falseToTrue[0].GetId())
	require.Equal(t, checker.ERR, falseToTrue[0].GetLevel())

	require.Empty(t, exclusiveBoolChanges(t, nil, boolPtr(false)),
		"adding exclusiveMaximum: false changes nothing")

	trueToNil := exclusiveBoolChanges(t, boolPtr(true), nil)
	require.Len(t, trueToNil, 1)
	require.Equal(t, "request-property-exclusive-max-unset", trueToNil[0].GetId())
	require.Equal(t, checker.INFO, trueToNil[0].GetLevel())
}
