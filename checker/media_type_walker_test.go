package checker_test

import (
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// helper: build a *diff.MethodDiff with a non-nil Revision operation so that
// NewApiChange (called via ctx.NewChange) can read OperationID.
func methodDiffWithRevision(opID string) *diff.MethodDiff {
	return &diff.MethodDiff{
		Revision: &openapi3.Operation{OperationID: opID},
	}
}

func TestWalkModifiedRequestBodySchemas_VisitsEachMediaType(t *testing.T) {
	op := methodDiffWithRevision("updatePets")
	op.RequestBodyDiff = &diff.RequestBodyDiff{
		ContentDiff: &diff.ContentDiff{
			MediaTypeModified: diff.ModifiedMediaTypes{
				"application/json": &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
				"application/xml":  &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
			},
		},
	}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/pets": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{"POST": op},
					},
				},
			},
		},
	}

	var visited []string
	var details []string
	checker.WalkModifiedRequestBodySchemas(d, &diff.OperationsSourcesMap{}, &checker.Config{}, func(ctx checker.MediaTypeChangeCtx) {
		visited = append(visited, ctx.MediaType)
		details = append(details, ctx.MediaTypeDetails)
		require.Equal(t, "/pets", ctx.Path)
		require.Equal(t, "POST", ctx.Method)
		require.NotNil(t, ctx.SchemaDiff)
		require.Empty(t, ctx.ResponseStatus, "request-body walker must leave ResponseStatus empty")
	})

	sort.Strings(visited)
	require.Equal(t, []string{"application/json", "application/xml"}, visited)
	// Two media types means formatMediaTypeDetails returns the qualifier on each.
	sort.Strings(details)
	require.Equal(t, []string{"(media type: application/json)", "(media type: application/xml)"}, details)
}

func TestWalkModifiedRequestBodySchemas_SkipsNilSchemaDiff(t *testing.T) {
	op := methodDiffWithRevision("updatePets")
	op.RequestBodyDiff = &diff.RequestBodyDiff{
		ContentDiff: &diff.ContentDiff{
			MediaTypeModified: diff.ModifiedMediaTypes{
				"application/json": &diff.MediaTypeDiff{SchemaDiff: nil},
				"application/xml":  &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
			},
		},
	}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/pets": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{"POST": op},
					},
				},
			},
		},
	}

	var visited []string
	checker.WalkModifiedRequestBodySchemas(d, &diff.OperationsSourcesMap{}, &checker.Config{}, func(ctx checker.MediaTypeChangeCtx) {
		visited = append(visited, ctx.MediaType)
	})
	require.Equal(t, []string{"application/xml"}, visited)
}

func TestWalkModifiedRequestBodySchemas_HandlesEmptyDiff(t *testing.T) {
	calls := 0
	processor := func(_ checker.MediaTypeChangeCtx) { calls++ }

	checker.WalkModifiedRequestBodySchemas(nil, &diff.OperationsSourcesMap{}, &checker.Config{}, processor)
	checker.WalkModifiedRequestBodySchemas(&diff.Diff{}, &diff.OperationsSourcesMap{}, &checker.Config{}, processor)
	checker.WalkModifiedRequestBodySchemas(&diff.Diff{PathsDiff: &diff.PathsDiff{}}, &diff.OperationsSourcesMap{}, &checker.Config{}, processor)

	require.Zero(t, calls, "walker must not invoke the processor when the diff is empty")
}

func TestWalkModifiedResponseSchemas_VisitsEachMediaTypeWithStatus(t *testing.T) {
	op := methodDiffWithRevision("getPet")
	op.ResponsesDiff = &diff.ResponsesDiff{
		Modified: diff.ModifiedResponses{
			"200": &diff.ResponseDiff{
				ContentDiff: &diff.ContentDiff{
					MediaTypeModified: diff.ModifiedMediaTypes{
						"application/json": &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
					},
				},
			},
			"404": &diff.ResponseDiff{
				ContentDiff: &diff.ContentDiff{
					MediaTypeModified: diff.ModifiedMediaTypes{
						"application/json": &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
					},
				},
			},
		},
	}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/pets/{id}": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{"GET": op},
					},
				},
			},
		},
	}

	var statuses []string
	checker.WalkModifiedResponseSchemas(d, &diff.OperationsSourcesMap{}, &checker.Config{}, func(ctx checker.MediaTypeChangeCtx) {
		statuses = append(statuses, ctx.ResponseStatus)
		require.Equal(t, "/pets/{id}", ctx.Path)
		require.Equal(t, "GET", ctx.Method)
		require.NotEmpty(t, ctx.ResponseStatus, "response walker must populate ResponseStatus")
		require.NotNil(t, ctx.SchemaDiff)
	})

	sort.Strings(statuses)
	require.Equal(t, []string{"200", "404"}, statuses)
}

// ctx.NewChange must pre-fill the plumbing (id, args, comment, operation/
// method/path) and attach the media-type detail string, so callers only
// need to chain check-specific extras like WithSources.
func TestMediaTypeChangeCtx_NewChange_PreFillsPlumbingAndDetails(t *testing.T) {
	op := methodDiffWithRevision("updatePets")
	op.RequestBodyDiff = &diff.RequestBodyDiff{
		ContentDiff: &diff.ContentDiff{
			MediaTypeModified: diff.ModifiedMediaTypes{
				"application/json": &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
				"application/xml":  &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
			},
		},
	}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/pets": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{"POST": op},
					},
				},
			},
		},
	}

	var changes []checker.ApiChange
	checker.WalkModifiedRequestBodySchemas(d, &diff.OperationsSourcesMap{}, allChecksConfig(), func(ctx checker.MediaTypeChangeCtx) {
		changes = append(changes, ctx.NewChange("request-body-any-of-added", []any{"#/components/schemas/Rabbit"}, ""))
	})
	require.Len(t, changes, 2)

	for _, c := range changes {
		require.Equal(t, "request-body-any-of-added", c.Id)
		require.Equal(t, []any{"#/components/schemas/Rabbit"}, c.Args)
		require.Equal(t, "POST", c.Operation)
		require.Equal(t, "/pets", c.Path)
		require.Equal(t, "updatePets", c.OperationId)
	}

	// Each media type must have its own detail string so a reader can tell
	// the two emitted changes apart; this is the bug class the helper
	// closes by construction.
	var details []string
	for _, c := range changes {
		details = append(details, c.Details)
	}
	sort.Strings(details)
	require.Equal(t, []string{"(media type: application/json)", "(media type: application/xml)"}, details)
}

// ctx.WalkProperties must visit every modified property in the schema diff
// and deliver a PropertyChangeCtx whose embedded MediaTypeChangeCtx carries
// the same plumbing as the outer ctx (so p.NewChange via field promotion
// produces an ApiChange with the right config / operation / method / path /
// MediaTypeDetails).
func TestMediaTypeChangeCtx_WalkProperties_VisitsEveryProperty(t *testing.T) {
	op := methodDiffWithRevision("updatePets")
	op.RequestBodyDiff = &diff.RequestBodyDiff{
		ContentDiff: &diff.ContentDiff{
			MediaTypeModified: diff.ModifiedMediaTypes{
				"application/json": &diff.MediaTypeDiff{
					SchemaDiff: &diff.SchemaDiff{
						PropertiesDiff: &diff.SchemasDiff{
							Modified: diff.ModifiedSchemasMap{
								"name": &diff.SchemaDiff{},
								"age":  &diff.SchemaDiff{},
							},
						},
					},
				},
			},
		},
	}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/pets": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{"POST": op},
					},
				},
			},
		},
	}

	var properties []string
	checker.WalkModifiedRequestBodySchemas(d, &diff.OperationsSourcesMap{}, allChecksConfig(), func(ctx checker.MediaTypeChangeCtx) {
		ctx.WalkProperties(func(p checker.PropertyChangeCtx) {
			properties = append(properties, p.PropertyName)
			require.NotNil(t, p.PropertyDiff)
			// The embedded ctx must carry the same media-type-level plumbing.
			require.Equal(t, "/pets", p.Path)
			require.Equal(t, "POST", p.Method)
			require.Equal(t, "application/json", p.MediaType)
		})
	})

	sort.Strings(properties)
	require.Equal(t, []string{"age", "name"}, properties)
}

// p.NewChange must produce an ApiChange equivalent to ctx.NewChange (via
// field promotion of MediaTypeChangeCtx). This is the bug-class prevention
// at property level: callers cannot accidentally drop the media-type details
// on a property-level emission.
func TestPropertyChangeCtx_NewChange_DelegatesToMediaTypeChangeCtx(t *testing.T) {
	op := methodDiffWithRevision("updatePets")
	op.RequestBodyDiff = &diff.RequestBodyDiff{
		ContentDiff: &diff.ContentDiff{
			MediaTypeModified: diff.ModifiedMediaTypes{
				"application/json": &diff.MediaTypeDiff{
					SchemaDiff: &diff.SchemaDiff{
						PropertiesDiff: &diff.SchemasDiff{
							Modified: diff.ModifiedSchemasMap{
								"role": &diff.SchemaDiff{},
							},
						},
					},
				},
				"application/xml": &diff.MediaTypeDiff{
					SchemaDiff: &diff.SchemaDiff{
						PropertiesDiff: &diff.SchemasDiff{
							Modified: diff.ModifiedSchemasMap{
								"role": &diff.SchemaDiff{},
							},
						},
					},
				},
			},
		},
	}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/pets": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{"POST": op},
					},
				},
			},
		},
	}

	var details []string
	checker.WalkModifiedRequestBodySchemas(d, &diff.OperationsSourcesMap{}, allChecksConfig(), func(ctx checker.MediaTypeChangeCtx) {
		ctx.WalkProperties(func(p checker.PropertyChangeCtx) {
			c := p.NewChange("request-property-any-of-added", []any{"#/components/schemas/Rabbit", p.PropertyName}, "")
			require.Equal(t, "POST", c.Operation)
			require.Equal(t, "/pets", c.Path)
			require.Equal(t, "updatePets", c.OperationId)
			details = append(details, c.Details)
		})
	})

	// One emission per (media type) since each media type has the same
	// single modified property; details must distinguish them.
	sort.Strings(details)
	require.Equal(t, []string{"(media type: application/json)", "(media type: application/xml)"}, details)
}

// Single media type means formatMediaTypeDetails returns the empty string;
// the helper must still produce a valid ApiChange (Details just empty).
func TestMediaTypeChangeCtx_NewChange_SingleMediaTypeHasEmptyDetails(t *testing.T) {
	op := methodDiffWithRevision("updatePets")
	op.RequestBodyDiff = &diff.RequestBodyDiff{
		ContentDiff: &diff.ContentDiff{
			MediaTypeModified: diff.ModifiedMediaTypes{
				"application/json": &diff.MediaTypeDiff{SchemaDiff: &diff.SchemaDiff{}},
			},
		},
	}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/pets": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{"POST": op},
					},
				},
			},
		},
	}

	var change checker.ApiChange
	checker.WalkModifiedRequestBodySchemas(d, &diff.OperationsSourcesMap{}, allChecksConfig(), func(ctx checker.MediaTypeChangeCtx) {
		change = ctx.NewChange("request-body-any-of-added", []any{"x"}, "")
	})

	require.Equal(t, "request-body-any-of-added", change.Id)
	require.Empty(t, change.Details, "single media type must not produce a (media type: ...) qualifier")
}
