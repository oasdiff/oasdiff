package checker

import (
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/internal/populatetest"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// setBoundSample populates the schema field whose json tag is the keyword,
// deriving the sample from the field's type; the keyword string is the only
// input, so a table row cannot need a hand-written sample.
func setBoundSample(t *testing.T, s *openapi3.Schema, keyword string) {
	t.Helper()
	typ := reflect.TypeFor[openapi3.Schema]()
	for i := range typ.NumField() {
		if name, _, _ := strings.Cut(typ.Field(i).Tag.Get("json"), ","); name == keyword {
			require.True(t, populatetest.NonZero(reflect.ValueOf(s).Elem().Field(i), keyword), keyword)
			return
		}
	}
	t.Fatalf("no openapi3.Schema field with json tag %q", keyword)
}

// keywordDoc builds a spec whose request or response schema carries the
// keyword sample when withKeyword is true, at body or property level.
func keywordDoc(t *testing.T, direction Direction, scope string, spec keywordSpec, withKeyword bool) *load.SpecInfo {
	node := &openapi3.Schema{}
	if withKeyword {
		setBoundSample(t, node, spec.keyword)
	}
	carrier := node
	if scope == "property" {
		carrier = &openapi3.Schema{
			Properties: openapi3.Schemas{"p": &openapi3.SchemaRef{Value: node}},
		}
	}

	plain := func() *openapi3.SchemaRef { return &openapi3.SchemaRef{Value: &openapi3.Schema{}} }
	requestSchema, responseSchema := plain(), plain()
	if direction == DirectionResponse {
		responseSchema = &openapi3.SchemaRef{Value: carrier}
	} else {
		requestSchema = &openapi3.SchemaRef{Value: carrier}
	}

	op := &openapi3.Operation{
		RequestBody: &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{
			Content: openapi3.Content{"application/json": &openapi3.MediaType{Schema: requestSchema}},
		}},
		Responses: openapi3.NewResponses(openapi3.WithStatus(200, &openapi3.ResponseRef{Value: &openapi3.Response{
			Description: new("ok"),
			Content:     openapi3.Content{"application/json": &openapi3.MediaType{Schema: responseSchema}},
		}})),
	}

	return &load.SpecInfo{Spec: &openapi3.T{
		OpenAPI: "3.1.0",
		Info:    &openapi3.Info{Title: "t", Version: "1.0.0"},
		Paths:   openapi3.NewPaths(openapi3.WithPath("/t", &openapi3.PathItem{Post: op})),
	}}
}

// Every generated rule fires: for each keyword, direction, scope, and
// action, a spec pair built from the table's own sample produces a change
// with the generated id at the rule's registered level, and the message and
// comment render as text rather than as their keys. The table that generates
// the rules also drives this gate, so a row cannot be registered untested.
// Every keywordSpecs row resolves to a diff.SchemaBound; the keyword string
// is the join between the checker table and the diff's bound registry.
func TestKeywordSpecsResolve(t *testing.T) {
	for _, spec := range keywordSpecs {
		_, ok := schemaBound(spec.keyword)
		require.True(t, ok, "%s does not resolve to a diff.SchemaBound", spec.keyword)
	}
}

func TestKeywordRulesFire(t *testing.T) {
	byId := map[string]BackwardCompatibilityRule{}
	for _, rule := range keywordRules() {
		byId[rule.Id] = rule
	}
	require.Len(t, byId, 66)

	localizer := NewDefaultLocalizer()
	config := NewConfig(GetAllChecks(),
		WithSingleCheck(KeywordSetUnsetCheck),
		WithSeverityLevels(map[string]Level{
			APIVersionNotBumpedId:      NONE,
			APIVersionDecreasedId:      NONE,
			APIMajorVersionNotBumpedId: NONE,
		}))

	for _, spec := range keywordSpecs {
		for _, direction := range []Direction{DirectionRequest, DirectionResponse} {
			for _, scope := range []string{"body", "property"} {
				for _, action := range keywordActions {
					id := keywordRuleId(direction, scope, spec.idName, action.action)
					rule, generated := byId[id]
					if !generated {
						continue
					}

					absent := keywordDoc(t, direction, scope, spec, false)
					present := keywordDoc(t, direction, scope, spec, true)
					base, revision := absent, present
					if action.action == "unset" {
						base, revision = present, absent
					}

					d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
					require.NoError(t, err, id)
					changes := CheckBackwardCompatibilityUntilLevel(config, d, osm, INFO)

					require.Len(t, changes, 1, id)
					change := changes[0]
					require.Equal(t, id, change.GetId())
					require.Equal(t, rule.Level, change.GetLevel(), id)
					text := change.GetUncolorizedText(localizer)
					require.NotContains(t, text, id, "message must render, not echo its key: %s", text)
					require.NotContains(t, change.GetComment(localizer), "-comment", id)
				}
			}
		}
	}
}
