package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// refLessCyclicDoc builds a spec whose request body property points at a
// schema that references itself directly, with no $ref. Such schemas cannot
// come from a document (a ref-less cycle has no serialized form); they are
// built in memory, e.g. by --flatten-allof merging a recursive component
// into an allOf overlay, or by library callers.
func refLessCyclicDoc(description string) *openapi3.T {
	node := &openapi3.Schema{
		Type:        &openapi3.Types{"object"},
		Description: description,
	}
	node.Properties = openapi3.Schemas{
		"filters": &openapi3.SchemaRef{Value: &openapi3.Schema{
			Type:  &openapi3.Types{"array"},
			Items: &openapi3.SchemaRef{Value: node},
		}},
	}

	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "1"},
		Paths: openapi3.NewPaths(openapi3.WithPath("/x", &openapi3.PathItem{
			Post: &openapi3.Operation{
				RequestBody: &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{
					Content: openapi3.Content{"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
							Type:       &openapi3.Types{"object"},
							Properties: openapi3.Schemas{"tree": &openapi3.SchemaRef{Value: node}},
						}},
					}},
				}},
				Responses: openapi3.NewResponses(openapi3.WithStatus(200, &openapi3.ResponseRef{
					Value: &openapi3.Response{Description: new("ok")},
				})),
			},
		})),
	}
}

// The circular-ref guard keys on Ref strings, so a ref-less cycle is
// invisible to it; the in-flight schema-pair guard cuts the cycle, and
// real changes must still be detected.
func TestRefLessCircularSchema(t *testing.T) {
	d, err := diff.Get(diff.NewConfig(), refLessCyclicDoc("recursive filter tree"), refLessCyclicDoc("recursive filter tree, modified"))
	require.NoError(t, err)

	// the description change on the cyclic node is still detected
	treeDiff := d.PathsDiff.Modified["/x"].OperationsDiff.Modified["POST"].
		RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].
		SchemaDiff.PropertiesDiff.Modified["tree"]
	require.NotNil(t, treeDiff)
	require.NotNil(t, treeDiff.DescriptionDiff)
	require.Equal(t, "recursive filter tree", treeDiff.DescriptionDiff.From)
	require.Equal(t, "recursive filter tree, modified", treeDiff.DescriptionDiff.To)
}
