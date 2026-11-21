package checker_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

type parameterFinding struct {
	id, operation string
}

func limitParameter(required bool, maximum *float64, enum ...any) *openapi3.Parameter {
	schema := openapi3.NewIntegerSchema()
	schema.Max = maximum
	schema.Enum = enum
	return &openapi3.Parameter{Name: "limit", In: openapi3.ParameterInQuery, Required: required, Schema: schema.NewRef()}
}

// parameterPositionsSpec puts param, when not nil, either on the /pets path
// item or on each of its operations.
func parameterPositionsSpec(t *testing.T, param *openapi3.Parameter, pathLevel bool) *load.SpecInfo {
	t.Helper()
	s, err := open("../data/checker/parameter_positions.yaml")
	require.NoError(t, err)
	if param == nil {
		return s
	}
	pathItem := s.Spec.Paths.Value("/pets")
	if pathLevel {
		pathItem.Parameters = openapi3.Parameters{{Value: param}}
	} else {
		for _, operation := range pathItem.Operations() {
			operation.Parameters = openapi3.Parameters{{Value: param}}
		}
	}
	return s
}

func parameterFindings(t *testing.T, base, revision *load.SpecInfo) []parameterFinding {
	t.Helper()
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	result := []parameterFinding{}
	for _, change := range checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO) {
		result = append(result, parameterFinding{change.GetId(), change.GetOperation()})
	}
	slices.SortFunc(result, func(a, b parameterFinding) int {
		return strings.Compare(a.id+a.operation, b.id+b.operation)
	})
	return result
}

// A request to an operation carries its own parameters and its path item's,
// so a change to a parameter gives the same findings whichever level it is
// declared at, and moving it between levels gives none.
func TestRequestParameterPositions(t *testing.T) {
	ten := float64(10)
	params := map[string]func() *openapi3.Parameter{
		"none":            func() *openapi3.Parameter { return nil },
		"optional":        func() *openapi3.Parameter { return limitParameter(false, nil) },
		"required":        func() *openapi3.Parameter { return limitParameter(true, nil) },
		"maximum 10":      func() *openapi3.Parameter { return limitParameter(false, &ten) },
		"enum 1, 2":       func() *openapi3.Parameter { return limitParameter(false, nil, 1, 2) },
		"enum 1, 2, 3":    func() *openapi3.Parameter { return limitParameter(false, nil, 1, 2, 3) },
		"required max 10": func() *openapi3.Parameter { return limitParameter(true, &ten) },
	}
	both := func(id string) []parameterFinding {
		return []parameterFinding{{id, "GET"}, {id, "POST"}}
	}

	tests := []struct {
		base, revision string
		// want is the findings with the parameter on the operations; on the
		// path item an added parameter has its own id
		want, wantPathLevel []parameterFinding
	}{
		{"none", "optional", both(checker.NewOptionalRequestParameterId), both(checker.NewOptionalRequestDefaultParameterToExistingPathId)},
		{"none", "required", both(checker.NewRequiredRequestParameterId), both(checker.NewRequiredRequestDefaultParameterToExistingPathId)},
		{"optional", "none", both(checker.RequestParameterRemovedId), nil},
		{"optional", "required", both(checker.RequestParameterBecomeRequiredId), nil},
		{"required", "optional", both(checker.RequestParameterBecomeOptionalId), nil},
		{"optional", "maximum 10", both(checker.RequestParameterMaxSetId), nil},
		{"enum 1, 2, 3", "enum 1, 2", both(checker.RequestParameterEnumValueRemovedId), nil},
		{"enum 1, 2", "enum 1, 2, 3", both(checker.RequestParameterEnumValueAddedId), nil},
		{"optional", "optional", []parameterFinding{}, nil},
		{"required max 10", "required max 10", []parameterFinding{}, nil},
	}

	for _, tc := range tests {
		t.Run(tc.base+" to "+tc.revision, func(t *testing.T) {
			for _, baseOnPath := range []bool{false, true} {
				for _, revisionOnPath := range []bool{false, true} {
					want := tc.want
					if revisionOnPath && tc.wantPathLevel != nil {
						want = tc.wantPathLevel
					}
					got := parameterFindings(t,
						parameterPositionsSpec(t, params[tc.base](), baseOnPath),
						parameterPositionsSpec(t, params[tc.revision](), revisionOnPath))
					require.Equal(t, want, got, "base on path item: %v, revision on path item: %v", baseOnPath, revisionOnPath)
				}
			}
		})
	}
}

// the operation's own limit overrides the path item's, so a change to the
// path item's limit doesn't reach the operation
func TestRequestParameterOverriddenByOperation(t *testing.T) {
	ten := float64(10)
	spec := func(pathLevel *openapi3.Parameter) *load.SpecInfo {
		s := parameterPositionsSpec(t, pathLevel, true)
		s.Spec.Paths.Value("/pets").Get.Parameters = openapi3.Parameters{{Value: limitParameter(true, nil)}}
		return s
	}
	require.Equal(t,
		[]parameterFinding{{checker.RequestParameterMaxSetId, "POST"}},
		parameterFindings(t, spec(limitParameter(false, nil)), spec(limitParameter(false, &ten))))
}

func TestRequestParameterWebhookPathLevel(t *testing.T) {
	spec := func(required bool) *load.SpecInfo {
		s := parameterPositionsSpec(t, nil, false)
		s.Spec.Webhooks["newPet"].Parameters = openapi3.Parameters{{Value: limitParameter(required, nil)}}
		return s
	}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), spec(false), spec(true))
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterRequiredValueUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterBecomeRequiredId, errs[0].GetId())
	require.Equal(t, "webhook:newPet", errs[0].GetPath())
	require.Equal(t, "POST", errs[0].GetOperation())
}

// an operation below the stability level is out of scope for a change to its
// path item's parameters too
func TestRequestParameterPathLevelStability(t *testing.T) {
	spec := func(required bool) *load.SpecInfo {
		s := parameterPositionsSpec(t, limitParameter(required, nil), true)
		s.Spec.Paths.Value("/pets").Get.Extensions = map[string]any{"x-stability-level": "alpha"}
		return s
	}
	require.Equal(t,
		[]parameterFinding{{checker.RequestParameterBecomeRequiredId, "POST"}},
		parameterFindings(t, spec(false), spec(true)))
}

// the source of a change to a path-level parameter is the parameter on the path item
func TestRequestParameterPathLevelSource(t *testing.T) {
	const file = "../data/checker/common_request_parameter_max_items_updated_base.yaml"
	s1, err := open(file, newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open(file, newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2.Spec.Paths.Value("/api/v1.0/groups").Parameters = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterRemovedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterRemovedId, errs[0].GetId())
	require.Equal(t, file, errs[0].GetBaseSource().File)
	require.Equal(t, 10, errs[0].GetBaseSource().Line)
}
