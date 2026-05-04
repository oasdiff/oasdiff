package allof_test

import (
	"context"
	"testing"

	"github.com/oasdiff/oasdiff/flatten/allof"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// exclusiveBoundBool creates an ExclusiveBound with a boolean value (OpenAPI 3.0 style).
func exclusiveBoundBool(b bool) openapi3.ExclusiveBound {
	return openapi3.ExclusiveBound{Bool: &b}
}

// exclusiveBoundValue creates an ExclusiveBound with a numeric value (OpenAPI 3.1 style).
func exclusiveBoundValue(v float64) openapi3.ExclusiveBound {
	return openapi3.ExclusiveBound{Value: &v}
}

// Single-schema round-trip tests: Merge on a schema with no allOf must not
// silently drop scalar fields. mergeInternal hand-copied a curated allowlist
// of base fields into result, so any field not on the list (Example,
// Deprecated, AllowEmptyValue, XML, ExternalDocs, Extensions) was lost on
// every Merge call — including pure passthroughs with no actual merging.
// These tests pin "round-trip preserves the field" so a future regression
// in the copy block is caught.

func TestMerge_SingleSchema_PreservesExample(t *testing.T) {
	merged, err := allof.Merge(openapi3.SchemaRef{
		Value: &openapi3.Schema{Example: "sample-value"},
	})
	require.NoError(t, err)
	require.Equal(t, "sample-value", merged.Example)
}

func TestMerge_SingleSchema_PreservesDeprecated(t *testing.T) {
	merged, err := allof.Merge(openapi3.SchemaRef{
		Value: &openapi3.Schema{Deprecated: true},
	})
	require.NoError(t, err)
	require.True(t, merged.Deprecated)
}

func TestMerge_SingleSchema_PreservesAllowEmptyValue(t *testing.T) {
	merged, err := allof.Merge(openapi3.SchemaRef{
		Value: &openapi3.Schema{AllowEmptyValue: true},
	})
	require.NoError(t, err)
	require.True(t, merged.AllowEmptyValue)
}

func TestMerge_SingleSchema_PreservesExternalDocs(t *testing.T) {
	docs := &openapi3.ExternalDocs{Description: "more info", URL: "https://example.com/docs"}
	merged, err := allof.Merge(openapi3.SchemaRef{
		Value: &openapi3.Schema{ExternalDocs: docs},
	})
	require.NoError(t, err)
	require.Same(t, docs, merged.ExternalDocs)
}

func TestMerge_SingleSchema_PreservesXML(t *testing.T) {
	xml := &openapi3.XML{Name: "user", Namespace: "https://example.com/ns"}
	merged, err := allof.Merge(openapi3.SchemaRef{
		Value: &openapi3.Schema{XML: xml},
	})
	require.NoError(t, err)
	require.Same(t, xml, merged.XML)
}

func TestMerge_SingleSchema_PreservesExtensions(t *testing.T) {
	ext := map[string]any{"x-internal-id": "abc-123"}
	merged, err := allof.Merge(openapi3.SchemaRef{
		Value: &openapi3.Schema{Extensions: ext},
	})
	require.NoError(t, err)
	require.Equal(t, ext, merged.Extensions)
}

// Single-schema with Type set must round-trip the Type. Pins regression
// against the duplicate `result.Value.Type = base.Value.Type` lines that
// existed before — if both were removed accidentally, the field would
// be lost; if only the duplicate was removed (the intended cleanup),
// this test still passes.
func TestMerge_SingleSchema_PreservesType(t *testing.T) {
	merged, err := allof.Merge(openapi3.SchemaRef{
		Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
	})
	require.NoError(t, err)
	require.NotNil(t, merged.Type)
	require.True(t, merged.Type.Is("string"))
}

// identical Default fields are merged successfully
func TestMerge_Default(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Default: 10,
			},
		})

	require.NoError(t, err)
	require.Equal(t, 10, merged.Default)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
						},
					},
				},
			},
		})

	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, nil, merged.Default)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Default: 10,
						},
					},
				},
			},
		})

	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, 10, merged.Default)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Default: 10,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Default: 10,
						},
					},
				},
			},
		})

	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, 10, merged.Default)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Default: "abc",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Default: "abc",
						},
					},
				},
			},
		})

	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, "abc", merged.Default)
}

// Conflicting Default values cannot be resolved
func TestMerge_DefaultFailure(t *testing.T) {
	_, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Default: 10,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Default: "abc",
						},
					},
				},
			},
		})

	require.EqualError(t, err, allof.DefaultErrorMessage)
}

// verify that if all ReadOnly fields are set to false, then the ReadOnly field in the merged schema is false.
func TestMerge_ReadOnlyIsSetToFalse(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							ReadOnly: false,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							ReadOnly: false,
						},
					},
				},
			},
		})

	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, false, merged.ReadOnly)
}

// verify that if there exists a ReadOnly field which is true, then the ReadOnly field in the merged schema is true.
func TestMerge_ReadOnlyIsSetToTrue(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							ReadOnly: true,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							ReadOnly: false,
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, true, merged.ReadOnly)
}

// verify that if all WriteOnly fields are set to false, then the WriteOnly field in the merged schema is false.
func TestMerge_WriteOnlyIsSetToFalse(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							WriteOnly: false,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							WriteOnly: false,
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, false, merged.WriteOnly)
}

// verify that if there exists a WriteOnly field which is true, then the WriteOnly field in the merged schema is true.
func TestMerge_WriteOnlyIsSetToTrue(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							WriteOnly: true,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							WriteOnly: false,
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, true, merged.WriteOnly)
}

// verify that if all nullable fields are set to true, then the nullable field in the merged schema is true.
func TestMerge_NullableIsSetToTrue(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							Nullable: true,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							Nullable: true,
						},
					},
				},
				Nullable: true,
			}})
	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, true, merged.Nullable)
}

// verify that if there exists a nullable field which is false, then the nullable field in the merged schema is false.
func TestMerge_NullableIsSetToFalse(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							Nullable: false,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							Nullable: true,
						},
					},
				},
				Nullable: true,
			}})
	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.Equal(t, false, merged.Nullable)
}

func TestMerge_NestedAllOfInProperties(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Properties: openapi3.Schemas{
					"prop1": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							AllOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										MinProps: 10,
										MaxProps: new(uint64(40)),
									},
								},
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										MinProps: 5,
										MaxProps: new(uint64(25)),
									},
								},
							},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.Properties["prop1"].Value.AllOf)
	require.Equal(t, uint64(10), merged.Properties["prop1"].Value.MinProps)
	require.Equal(t, uint64(25), *merged.Properties["prop1"].Value.MaxProps)
}

func TestMerge_NestedAllOfInNot(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Not: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						AllOf: openapi3.SchemaRefs{
							&openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:     &openapi3.Types{"object"},
									MinProps: 10,
									MaxProps: new(uint64(40)),
								},
							},
							&openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:     &openapi3.Types{"object"},
									MinProps: 5,
									MaxProps: new(uint64(25)),
								},
							},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.Not.Value.AllOf)
	require.Equal(t, uint64(10), merged.Not.Value.MinProps)
	require.Equal(t, uint64(25), *merged.Not.Value.MaxProps)
}

func TestMerge_NestedAllOfInOneOf(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				OneOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							AllOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										MinProps: 10,
										MaxProps: new(uint64(40)),
									},
								},
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										MinProps: 5,
										MaxProps: new(uint64(25)),
									},
								},
							},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.OneOf[0].Value.AllOf)
	require.Equal(t, uint64(10), merged.OneOf[0].Value.MinProps)
	require.Equal(t, uint64(25), *merged.OneOf[0].Value.MaxProps)
}

// AllOf is empty in base schema, but there is nested non-empty AllOf in base schema.
func TestMerge_NestedAllOfInAnyOf(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AnyOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							AllOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										MinProps: 10,
										MaxProps: new(uint64(40)),
									},
								},
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										MinProps: 5,
										MaxProps: new(uint64(25)),
									},
								},
							},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.AnyOf[0].Value.AllOf)
	require.Equal(t, uint64(10), merged.AnyOf[0].Value.MinProps)
	require.Equal(t, uint64(25), *merged.AnyOf[0].Value.MaxProps)
}

// identical numeric types are merged successfully
func TestMerge_TypeNumeric(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"number"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"number"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.True(t, merged.Properties["prop1"].Value.Type.Is("number"))

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"integer"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"integer"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, &openapi3.Types{"integer"}, merged.Properties["prop1"].Value.Type)
}

// Conflicting numeric types are merged successfully
func TestMerge_TypeNumericConflictResolved(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"integer"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"number"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.True(t, merged.Properties["prop1"].Value.Type.Is("integer"))
}

// Conflicting types cannot be resolved
func TestMerge_TypeFailure(t *testing.T) {
	_, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"integer"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})

	require.EqualError(t, err, allof.TypeErrorMessage)
}

// if ExclusiveMax is true on the minimum Max value, then ExclusiveMax is true in the merged schema.
func TestMerge_ExclusiveMaxIsTrue(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMax: exclusiveBoundBool(true),
							Max:          new(1.0),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMax: exclusiveBoundBool(false),
							Max:          new(2.0),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.True(t, merged.ExclusiveMax.IsTrue())
}

// if ExclusiveMax is false on the minimum Max value, then ExclusiveMax is false in the merged schema.
func TestMerge_ExclusiveMaxIsFalse(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMax: exclusiveBoundBool(false),
							Max:          new(1.0),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMax: exclusiveBoundBool(true),
							Max:          new(2.0),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.False(t, merged.ExclusiveMax.IsTrue())
}

// if ExclusiveMin is false on the highest Min value, then ExclusiveMin is false in the merged schema.
func TestMerge_ExclusiveMinIsFalse(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMin: exclusiveBoundBool(false),
							Min:          new(40.0),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMin: exclusiveBoundBool(true),
							Min:          new(5.0),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.False(t, merged.ExclusiveMin.IsTrue())
}

// if ExclusiveMin is true on the highest Min value, then ExclusiveMin is true in the merged schema.
func TestMerge_ExclusiveMinIsTrue(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMin: exclusiveBoundBool(true),
							Min:          new(40.0),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:         &openapi3.Types{"object"},
							ExclusiveMin: exclusiveBoundBool(false),
							Min:          new(5.0),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.True(t, merged.ExclusiveMin.IsTrue())
}

// OpenAPI 3.1: two numeric `exclusiveMinimum` values (no `minimum` on either
// side). The merged schema must keep the higher one — failing this test was
// the original symptom in issue #868: both values were silently dropped.
func TestMerge_ExclusiveMinNumeric_PicksHigher(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMin: exclusiveBoundValue(10),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMin: exclusiveBoundValue(5),
					}},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.Min, "Min must be nil for 3.1 numeric exclusive form")
	require.NotNil(t, merged.ExclusiveMin.Value)
	require.Equal(t, 10.0, *merged.ExclusiveMin.Value)
}

// Same value on both sides — `minimum: 5` (inclusive) vs
// `exclusiveMinimum: 5` (strict). The strict bound rejects the value
// 5 itself and is therefore more restrictive at the boundary, so the
// merged result is `exclusiveMinimum: 5` with `minimum` cleared.
func TestMerge_ExclusiveMinNumeric_BeatsIdenticalInclusive(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type: &openapi3.Types{"number"},
						Min:  new(5.0),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMin: exclusiveBoundValue(5),
					}},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.Min)
	require.NotNil(t, merged.ExclusiveMin.Value)
	require.Equal(t, 5.0, *merged.ExclusiveMin.Value)
}

// Two subschemas: `minimum: 0` (inclusive) and `exclusiveMinimum: 5`
// (strict). The strict bound at 5 is more restrictive than the inclusive
// bound at 0, so the merged result is `exclusiveMinimum: 5` with
// `minimum` cleared.
func TestMerge_ExclusiveMinNumeric_BeatsLowerInclusive(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type: &openapi3.Types{"number"},
						Min:  new(0.0),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMin: exclusiveBoundValue(5),
					}},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.Min)
	require.NotNil(t, merged.ExclusiveMin.Value)
	require.Equal(t, 5.0, *merged.ExclusiveMin.Value)
}

// Two subschemas: `minimum: 10` (inclusive) and `exclusiveMinimum: 5`
// (strict). Even though `exclusiveMinimum: 5` is strict, its value is
// lower, so the inclusive `minimum: 10` is more restrictive overall.
// The merged result is `minimum: 10` with no `exclusiveMinimum`.
func TestMerge_InclusiveMinBeatsLowerNumericExclusive(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type: &openapi3.Types{"number"},
						Min:  new(10.0),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMin: exclusiveBoundValue(5),
					}},
				},
			}})
	require.NoError(t, err)
	require.NotNil(t, merged.Min)
	require.Equal(t, 10.0, *merged.Min)
	require.Nil(t, merged.ExclusiveMin.Value)
	require.False(t, merged.ExclusiveMin.IsTrue())
}

// One schema has both `minimum: 0` (inclusive) and `exclusiveMinimum: 5`
// (strict) — within that schema, the strict bound at 5 wins (effective
// constraint: x > 5). A second schema contributes `minimum: 10`
// (inclusive), which is more restrictive than the first schema's
// effective `> 5`. The merged result is `minimum: 10` with no
// `exclusiveMinimum`.
func TestMerge_ExclusiveMinNumeric_WithSiblingInclusive(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						Min:          new(0.0),
						ExclusiveMin: exclusiveBoundValue(5),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type: &openapi3.Types{"number"},
						Min:  new(10.0),
					}},
				},
			}})
	require.NoError(t, err)
	// Effective lower bounds: > 5 (from schema 1's numeric exclusive),
	// >= 0 (from schema 1's minimum), >= 10 (from schema 2). The most
	// restrictive is >= 10.
	require.NotNil(t, merged.Min)
	require.Equal(t, 10.0, *merged.Min)
	require.Nil(t, merged.ExclusiveMin.Value)
	require.False(t, merged.ExclusiveMin.IsTrue())
}

// Mirror cases for the upper bound.

func TestMerge_ExclusiveMaxNumeric_PicksLower(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMax: exclusiveBoundValue(100),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMax: exclusiveBoundValue(50),
					}},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.Max)
	require.NotNil(t, merged.ExclusiveMax.Value)
	require.Equal(t, 50.0, *merged.ExclusiveMax.Value)
}

func TestMerge_ExclusiveMaxNumeric_BeatsHigherInclusive(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type: &openapi3.Types{"number"},
						Max:  new(100.0),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMax: exclusiveBoundValue(50),
					}},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.Max)
	require.NotNil(t, merged.ExclusiveMax.Value)
	require.Equal(t, 50.0, *merged.ExclusiveMax.Value)
}

func TestMerge_InclusiveMaxBeatsHigherNumericExclusive(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type: &openapi3.Types{"number"},
						Max:  new(50.0),
					}},
					&openapi3.SchemaRef{Value: &openapi3.Schema{
						Type:         &openapi3.Types{"number"},
						ExclusiveMax: exclusiveBoundValue(100),
					}},
				},
			}})
	require.NoError(t, err)
	require.NotNil(t, merged.Max)
	require.Equal(t, 50.0, *merged.Max)
	require.Nil(t, merged.ExclusiveMax.Value)
	require.False(t, merged.ExclusiveMax.IsTrue())
}

// Per docs/ALLOF.md, source locations are not available for changes detected
// in flattened schemas — the merged schema is a new construct that doesn't
// correspond to a single line in the original file. This test pins that
// behavior so a future change to the merge code can't accidentally leak
// stale origin metadata (e.g. a `minimum` field origin pointing at the
// subschema whose value lost the merge), which would surface as wrong
// line numbers in downstream consumers.
func TestMerge_FlattenedNumericBounds_HaveNoSourceLocation(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`openapi: 3.1.0
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Bounded:
      allOf:
        - {type: number, minimum: 5}
        - {type: number, exclusiveMinimum: 6}
`))
	require.NoError(t, err)

	// Sanity-check: the loader actually populated origin info on the input —
	// otherwise the post-merge nil assertion would be vacuously true.
	require.NotNil(t, doc.Components.Schemas["Bounded"].Value.Origin,
		"input schema must have origin tracking populated for this test to be meaningful")

	merged, err := allof.Merge(*doc.Components.Schemas["Bounded"])
	require.NoError(t, err)

	// Sanity-check: merge picked the more restrictive bound (exclusive 6
	// beats inclusive 5).
	require.Nil(t, merged.Min)
	require.NotNil(t, merged.ExclusiveMin.Value)
	require.Equal(t, 6.0, *merged.ExclusiveMin.Value)

	// The actual assertion: no origin metadata on the merged schema.
	require.Nil(t, merged.Origin,
		"flattened schema must not carry origin metadata (per docs/ALLOF.md)")
}

// merge multiple Not inside AllOf
func TestMerge_Not(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Not: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
								},
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Not: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"integer"},
								},
							},
						},
					},
				},
			}})

	require.NoError(t, err)
	require.Equal(t, &openapi3.Types{"string"}, merged.Not.Value.AnyOf[0].Value.Type)
	require.True(t, merged.Not.Value.AnyOf[1].Value.Type.Is("integer"))
}

// merge multiple OneOf inside AllOf
func TestMerge_OneOf(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"prop1"},
									},
								},
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"prop2"},
									},
								},
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"prop2"},
									},
								},
							},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, []string{"prop1", "prop2"}, merged.OneOf[0].Value.Required)
	require.Equal(t, []string{"prop2"}, merged.OneOf[1].Value.Required)
}

// merge multiple AnyOf inside AllOf
func TestMerge_AnyOf(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							AnyOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"string"},
									},
								},
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"boolean"},
									},
								},
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							AnyOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"boolean"},
									},
								},
							},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, []string{"string", "boolean"}, merged.AnyOf[0].Value.Required)
}

// conflicting uniqueItems values are merged successfully
func TestMerge_UniqueItemsTrue(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							UniqueItems: true,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							UniqueItems: false,
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, true, merged.UniqueItems)
}

// non-conflicting uniqueItems values are merged successfully
func TestMerge_UniqueItemsFalse(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							UniqueItems: false,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							UniqueItems: false,
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, false, merged.UniqueItems)
}

// Item merge fails due to conflicting item types.
func TestMerge_Items_Failure(t *testing.T) {
	_, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"test": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"array"},
										Items: &openapi3.SchemaRef{
											Value: &openapi3.Schema{
												Type: &openapi3.Types{"integer"},
											},
										},
									},
								},
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"test": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"array"},
										Items: &openapi3.SchemaRef{
											Value: &openapi3.Schema{
												Type: &openapi3.Types{"string"},
											},
										},
									},
								},
							},
						},
					},
				},
			}})
	require.EqualError(t, err, allof.TypeErrorMessage)
}

// items are merged successfully when there are no conflicts
func TestMerge_Items(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"test": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"array"},
										Items: &openapi3.SchemaRef{
											Value: &openapi3.Schema{
												Type: &openapi3.Types{"integer"},
											},
										},
									},
								},
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"test": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"array"},
										Items: &openapi3.SchemaRef{
											Value: &openapi3.Schema{
												Type: &openapi3.Types{"integer"},
											},
										},
									},
								},
							},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Nil(t, merged.AllOf)
	require.True(t, merged.Properties["test"].Value.Type.Is("array"))
	require.True(t, merged.Properties["test"].Value.Items.Value.Type.Is("integer"))
}

func TestMerge_MultipleOfContained(t *testing.T) {

	//todo - more tests
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:       &openapi3.Types{"object"},
							MultipleOf: new(10.0),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:       &openapi3.Types{"object"},
							MultipleOf: new(2.0),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, float64(10), *merged.MultipleOf)
}

func TestMerge_MultipleOfDecimal(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:       &openapi3.Types{"object"},
							MultipleOf: new(11.0),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:       &openapi3.Types{"object"},
							MultipleOf: new(0.7),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, float64(77), *merged.MultipleOf)
}

func TestMerge_EnumContained(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Enum: []any{"1", nil, 1},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Enum: []any{"1"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.ElementsMatch(t, []any{"1"}, merged.Enum)
}

// enum merge fails if the intersection of enum values is empty.
func TestMerge_EnumNoIntersection(t *testing.T) {
	_, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Enum: []any{"1", nil},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Enum: []any{"2"},
						},
					},
				},
			}})
	require.Error(t, err)
}

// Properties range is the most restrictive
func TestMerge_RangeProperties(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							MinProps: 10,
							MaxProps: new(uint64(40)),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							MinProps: 5,
							MaxProps: new(uint64(25)),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, uint64(10), merged.MinProps)
	require.Equal(t, uint64(25), *merged.MaxProps)
}

// Items range is the most restrictive
func TestMerge_RangeItems(t *testing.T) {

	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							MinItems: 10,
							MaxItems: new(uint64(40)),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:     &openapi3.Types{"object"},
							MinItems: 5,
							MaxItems: new(uint64(25)),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, uint64(10), merged.MinItems)
	require.Equal(t, uint64(25), *merged.MaxItems)
}

func TestMerge_Range(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Min:  new(10.0),
							Max:  new(40.0),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Min:  new(5.0),
							Max:  new(25.0),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, float64(10), *merged.Min)
	require.Equal(t, float64(25), *merged.Max)
}

func TestMerge_MaxLength(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							MaxLength: new(uint64(10)),
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							MaxLength: new(uint64(20)),
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, uint64(10), *merged.MaxLength)
}

func TestMerge_MinLength(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							MinLength: 10,
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:      &openapi3.Types{"object"},
							MinLength: 20,
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, uint64(20), merged.MinLength)
}

func TestMerge_Description(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Description: "desc0",
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							Description: "desc1",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							Description: "desc2",
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, "desc0", merged.Description)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							Description: "desc1",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"object"},
							Description: "desc2",
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, "desc1", merged.Description)
}

// non-conflicting types are merged successfully
func TestMerge_NonConflictingType(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.True(t, merged.Type.Is("object"))
}

// schema cannot be merged if types are conflicting
func TestMerge_FailsOnConflictingTypes(t *testing.T) {
	_, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"name": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
									},
								},
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"name": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"object"},
									},
								},
							},
						},
					},
				},
			}})
	require.Error(t, err)
}

func TestMerge_Title(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Title: "base schema",
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:  &openapi3.Types{"object"},
							Title: "first",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:  &openapi3.Types{"object"},
							Title: "second",
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, "base schema", merged.Title)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:  &openapi3.Types{"object"},
							Title: "first",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:  &openapi3.Types{"object"},
							Title: "second",
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, "first", merged.Title)
}

// merge conflicting integer formats
func TestMerge_FormatInteger(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Format: allof.FormatInt32,
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Format: allof.FormatInt64,
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, allof.FormatInt32, merged.Properties["prop1"].Value.Format)
}

// merge conflicting float formats
func TestMerge_FormatFloat(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Format: allof.FormatFloat,
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Format: allof.FormatDouble,
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, allof.FormatFloat, merged.Properties["prop1"].Value.Format)
}

// merge conflicting integer and float formats
func TestMerge_NumericFormat(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Format: allof.FormatFloat,
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Format: allof.FormatDouble,
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Format: allof.FormatInt32,
									},
								},
							},
							Type: &openapi3.Types{"object"},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, allof.FormatInt32, merged.Properties["prop1"].Value.Format)
}

func TestMerge_Format(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:   &openapi3.Types{"object"},
							Format: "date",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:   &openapi3.Types{"object"},
							Format: "date",
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, "date", merged.Format)
}

func TestMerge_Format_Failure(t *testing.T) {
	_, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:   &openapi3.Types{"object"},
							Format: "date",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:   &openapi3.Types{"object"},
							Format: "byte",
						},
					},
				},
			}})
	require.EqualError(t, err, allof.FormatErrorMessage)
}

func TestMerge_EmptySchema(t *testing.T) {
	schema := openapi3.SchemaRef{
		Value: &openapi3.Schema{},
	}
	merged, err := allof.Merge(schema)
	require.NoError(t, err)
	require.Equal(t, schema.Value, merged)
}

func TestMerge_NoAllOf(t *testing.T) {
	schema := openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Title: "test",
		}}
	merged, err := allof.Merge(schema)
	require.NoError(t, err)
	require.Equal(t, schema.Value, merged)
}

func TestMerge_TwoObjects(t *testing.T) {

	obj1 := openapi3.Schemas{
		"description": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
		},
	}

	obj2 := openapi3.Schemas{
		"name": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
		},
	}

	schema := openapi3.SchemaRef{
		Value: &openapi3.Schema{
			AllOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:       &openapi3.Types{"object"},
						Properties: obj1,
					},
				},
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:       &openapi3.Types{"object"},
						Properties: obj2,
					},
				},
			},
		}}

	merged, err := allof.Merge(schema)
	require.NoError(t, err)
	require.Len(t, merged.AllOf, 0)
	require.Len(t, merged.Properties, 2)
	require.Equal(t, obj1["description"].Value.Type, merged.Properties["description"].Value.Type)
	require.Equal(t, obj2["name"].Value.Type, merged.Properties["name"].Value.Type)
}

func TestMerge_OneObjectOneProp(t *testing.T) {

	object := openapi3.Schemas{
		"description": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
		},
	}

	schema := openapi3.SchemaRef{
		Value: &openapi3.Schema{
			AllOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:       &openapi3.Types{"object"},
						Properties: object,
					},
				},
			},
		}}

	merged, err := allof.Merge(schema)
	require.NoError(t, err)
	require.Len(t, merged.Properties, 1)
	require.Equal(t, object["description"].Value.Type, merged.Properties["description"].Value.Type)
}

func TestMerge_OneObjectNoProps(t *testing.T) {

	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:       &openapi3.Types{"object"},
							Properties: openapi3.Schemas{},
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Len(t, merged.Properties, 0)
}

func TestMerge_OverlappingProps(t *testing.T) {

	obj1 := openapi3.Schemas{
		"description": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Title: "first",
			},
		},
	}

	obj2 := openapi3.Schemas{
		"description": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Title: "second",
			},
		},
	}

	schema := openapi3.SchemaRef{
		Value: &openapi3.Schema{
			AllOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:       &openapi3.Types{"object"},
						Properties: obj1,
					},
				},
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:       &openapi3.Types{"object"},
						Properties: obj2,
					},
				},
			},
		}}
	merged, err := allof.Merge(schema)
	require.NoError(t, err)
	require.Len(t, merged.AllOf, 0)
	require.Len(t, merged.Properties, 1)
	require.Equal(t, (*obj1["description"].Value), (*merged.Properties["description"].Value))
}

// if additionalProperties is false, then the merged additionalProperties is the intersection of relevant properties.
func TestMerge_AdditionalProperties_False(t *testing.T) {
	apFalse := false
	apTrue := true

	var firstPropEnum []any
	var secondPropEnum []any
	var thirdPropEnum []any

	firstPropEnum = append(firstPropEnum, "1", "5", "3")
	secondPropEnum = append(secondPropEnum, "1", "8", "7")
	thirdPropEnum = append(thirdPropEnum, "3", "8", "5")

	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Enum: firstPropEnum,
									},
								},
								"name": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
									},
								},
							},
							AdditionalProperties: openapi3.BoolSchema{
								Has: &apTrue,
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"prop2": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Enum: secondPropEnum,
									},
								},
							},
							AdditionalProperties: openapi3.BoolSchema{
								Has: &apFalse,
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"prop2": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Enum: thirdPropEnum,
									},
								},
							},
							AdditionalProperties: openapi3.BoolSchema{
								Has: &apFalse,
							},
						},
					},
				}}})
	require.NoError(t, err)
	require.Equal(t, "8", merged.Properties["prop2"].Value.Enum[0])
}

// if additionalProperties is true, then the merged additionalProperties is the intersection of all properties.
func TestMerge_AdditionalProperties_True(t *testing.T) {
	apTrue := true

	var firstPropEnum []any
	var secondPropEnum []any
	var thirdPropEnum []any

	firstPropEnum = append(firstPropEnum, "1", "5", "3")
	secondPropEnum = append(secondPropEnum, "1", "8", "7")
	thirdPropEnum = append(thirdPropEnum, "3", "8", "5")

	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"prop1": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Enum: firstPropEnum,
									},
								},
								"name": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
									},
								},
							},
							AdditionalProperties: openapi3.BoolSchema{
								Has: &apTrue,
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"prop2": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Enum: secondPropEnum,
									},
								},
							},
							AdditionalProperties: openapi3.BoolSchema{
								Has: &apTrue,
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: openapi3.Schemas{
								"prop2": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Enum: thirdPropEnum,
									},
								},
							},
							AdditionalProperties: openapi3.BoolSchema{
								Has: &apTrue,
							},
						},
					},
				}}})
	require.NoError(t, err)
	require.True(t, merged.Properties["name"].Value.Type.Is("string"))
}

func TestMergeAllOf_Pattern(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Pattern: "abc",
			}})
	require.NoError(t, err)
	require.Equal(t, "abc", merged.Pattern)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Pattern: "abc",
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, "abc", merged.Pattern)

	merged, err = allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Pattern: "foo",
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    &openapi3.Types{"object"},
							Pattern: "bar",
						},
					},
				},
			}})
	require.NoError(t, err)
	require.Equal(t, "(?=foo)(?=bar)", merged.Pattern)
}

func TestMerge_Required(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		filename string
	}{
		{"testdata/properties.yml"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Load in the reference spec from the testdata
			sl := openapi3.NewLoader()
			sl.IsExternalRefsAllowed = true
			doc, err := sl.LoadFromFile(test.filename)
			require.NoError(t, err, "loading test file")
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating spec")
			merged, err := allof.Merge(*doc.Paths.Value("/products").Get.Responses.Value("200").Value.Content["application/json"].Schema)
			require.NoError(t, err)

			props := merged.Properties
			require.Len(t, props, 3)
			require.Contains(t, props, "id")
			require.Contains(t, props, "createdAt")
			require.Contains(t, props, "otherId")

			required := merged.Required
			require.Len(t, required, 2)
			require.Contains(t, required, "id")
			require.Contains(t, required, "otherId")
		})
	}
}

func TestMerge_CircularAllOf(t *testing.T) {
	doc := loadSpec(t, "testdata/circular1.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["AWSEnvironmentSettings"])
	require.NoError(t, err)
	require.Empty(t, merged.AllOf)
	require.Empty(t, merged.OneOf)

	require.True(t, merged.Properties["serviceEndpoints"].Value.Type.Is("string"))
	require.True(t, merged.Properties["region"].Value.Type.Is("string"))
}

// A single OneOf field is pruned if it references it's parent schema
func TestMerge_OneOfIsPruned(t *testing.T) {
	doc := loadSpec(t, "testdata/circular2.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["OneOf_Is_Pruned_B"])
	require.NoError(t, err)
	require.Empty(t, merged.AllOf)
	require.Empty(t, merged.OneOf)
}

// A single OneOf field is not pruned if it does not reference it's parent schema
func TestMerge_OneOfIsNotPruned(t *testing.T) {
	doc := loadSpec(t, "testdata/circular2.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["OneOf_Is_Not_Pruned_B"])
	require.NoError(t, err)
	require.Empty(t, merged.AllOf)
	require.NotEmpty(t, merged.OneOf)
}

// A single AnyOf field is pruned if it references it's parent schema
func TestMerge_AnyOfIsPruned(t *testing.T) {
	doc := loadSpec(t, "testdata/circular2.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["AnyOf_Is_Pruned_B"])
	require.NoError(t, err)
	require.Empty(t, merged.AllOf)
	require.Empty(t, merged.AnyOf)
}

// A single AnyOf field is not pruned if it does not reference it's parent schema
func TestMerge_AnyOfIsNotPruned(t *testing.T) {
	doc := loadSpec(t, "testdata/circular2.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["AnyOf_Is_Not_Pruned_B"])
	require.NoError(t, err)
	require.Empty(t, merged.AllOf)
	require.NotEmpty(t, merged.AnyOf)
}

func TestMerge_ComplexOneOfIsPruned(t *testing.T) {
	doc := loadSpec(t, "testdata/prune-oneof.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["SchemaWithWithoutOneOf"])
	require.NoError(t, err)
	require.Empty(t, merged.OneOf)
}

func TestMerge_ComplexOneOfIsNotPruned(t *testing.T) {
	doc := loadSpec(t, "testdata/prune-oneof.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["ThirdSchema"])
	require.NoError(t, err)
	require.NotEmpty(t, merged.OneOf)
	require.Len(t, merged.OneOf, 2)

	merged, err = allof.Merge(*doc.Components.Schemas["ComplexSchema"])
	require.NoError(t, err)
	require.NotEmpty(t, merged.OneOf)
	require.Len(t, merged.OneOf, 2)

	merged, err = allof.Merge(*doc.Components.Schemas["SchemaWithOneOf"])
	require.NoError(t, err)
	require.NotEmpty(t, merged.OneOf)
	require.Len(t, merged.OneOf, 2)
}

func TestMerge_SingleCircularItemInAllOf(t *testing.T) {
	doc := loadSpec(t, "testdata/single_circular_item.yaml")
	merged, err := allof.Merge(*doc.Components.Schemas["GameResult"])
	require.NoError(t, err)
	require.Empty(t, merged.Properties["MainCharacter"].Value.AllOf)
	require.Len(t, *merged.Properties["MainCharacter"].Value.Type, 1)
	require.Contains(t, *merged.Properties["MainCharacter"].Value.Type, "array")
	require.Equal(t, "#/components/schemas/GameResult", merged.Properties["MainCharacter"].Value.Items.Ref)
}

func loadSpec(t *testing.T, path string) *openapi3.T {
	ctx := context.Background()
	sl := openapi3.NewLoader()
	doc, err := sl.LoadFromFile(path)
	require.NoError(t, err, "loading test file")
	err = doc.Validate(ctx)
	require.NoError(t, err)
	return doc
}

// Test for issue #710: identical oneOf groups in allOf should be deduplicated
// instead of producing a cartesian product
func TestMerge_IdenticalOneOfGroups(t *testing.T) {
	// Create two allOf subschemas with identical oneOf groups
	oneOfA := &openapi3.SchemaRef{
		Ref: "#/components/schemas/A",
		Value: &openapi3.Schema{
			Type:     &openapi3.Types{"object"},
			Required: []string{"propA"},
		},
	}
	oneOfB := &openapi3.SchemaRef{
		Ref: "#/components/schemas/B",
		Value: &openapi3.Schema{
			Type:     &openapi3.Types{"object"},
			Required: []string{"propB"},
		},
	}

	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								oneOfA,
								oneOfB,
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								oneOfA,
								oneOfB,
							},
						},
					},
				},
			}})

	require.NoError(t, err)
	// Without the fix, this would produce 4 oneOf items (cartesian product)
	// With the fix, identical oneOf groups are deduplicated, resulting in 2 items
	require.Len(t, merged.OneOf, 2, "identical oneOf groups should be deduplicated, not produce cartesian product")
}

// Test that oneOf groups with same pointer references are deduplicated
func TestMerge_SamePointerOneOfGroups(t *testing.T) {
	// Create shared inline schemas (same pointer)
	inlineA := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:     &openapi3.Types{"object"},
			Required: []string{"propA"},
		},
	}
	inlineB := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:     &openapi3.Types{"object"},
			Required: []string{"propB"},
		},
	}

	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								inlineA,
								inlineB,
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								inlineA,
								inlineB,
							},
						},
					},
				},
			}})

	require.NoError(t, err)
	// Same pointer references should be deduplicated
	require.Len(t, merged.OneOf, 2, "oneOf groups with same pointer references should be deduplicated")
}

// Test that oneOf groups with refs in different order are still deduplicated (tests sorting)
func TestMerge_OneOfGroupsReversedOrder(t *testing.T) {
	oneOfA := &openapi3.SchemaRef{
		Ref: "#/components/schemas/A",
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
		},
	}
	oneOfB := &openapi3.SchemaRef{
		Ref: "#/components/schemas/B",
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
		},
	}

	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								oneOfA, // A first
								oneOfB,
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								oneOfB, // B first (reversed order)
								oneOfA,
							},
						},
					},
				},
			}})

	require.NoError(t, err)
	// Groups with same refs in different order should be deduplicated after sorting
	require.Len(t, merged.OneOf, 2, "oneOf groups with same refs in different order should be deduplicated")
}

// Test different oneOf groups are not deduplicated
func TestMerge_DifferentOneOfGroups(t *testing.T) {
	merged, err := allof.Merge(
		openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Ref: "#/components/schemas/A",
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"propA"},
									},
								},
							},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							OneOf: openapi3.SchemaRefs{
								&openapi3.SchemaRef{
									Ref: "#/components/schemas/B",
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"propB"},
									},
								},
							},
						},
					},
				},
			}})

	require.NoError(t, err)
	// Different oneOf groups should produce cartesian product (1 * 1 = 1 in this case)
	require.Len(t, merged.OneOf, 1)
}
