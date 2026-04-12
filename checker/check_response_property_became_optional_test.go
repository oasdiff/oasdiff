package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing required response property to optional
func TestResponsePropertyBecameOptionalCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_became_optional_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_became_optional_revision.yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameOptionalCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyBecameOptionalId,
		Args:        []any{"data/name", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_became_optional_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing required response property to optional — schema defined in an external $ref file.
// Verifies that source tracking (base source) points to the correct line in the $ref'd file.
// This exercises the yaml3 root-mapping origin fix: previously, schemas loaded from $ref'd files
// had Origin==nil because document() in yaml3 never injected __origin__ for the root mapping.
func TestResponsePropertyBecameOptionalCheck_ExternalRef(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	loader.IsExternalRefsAllowed = true

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/ref-chain-example/base/openapi.yaml"))
	require.NoError(t, err)
	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/ref-chain-example/revision/openapi.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameOptionalCheck), d, osm, checker.ERR)
	require.NotEmpty(t, errs)

	// Find an error for "id" becoming optional
	var found *checker.ApiChange
	for i := range errs {
		c, ok := errs[i].(checker.ApiChange)
		if ok && c.Id == checker.ResponsePropertyBecameOptionalId && len(c.Args) > 0 && c.Args[0] == "id" {
			found = &c
			break
		}
	}
	require.NotNil(t, found, "expected response-property-became-optional for 'id'")

	// Base source must point to the line of "- id" in the external schema file.
	// In data/ref-chain-example/base/schemas/pet.yaml, "- id" is at line 3.
	base := found.BaseSource
	require.NotNil(t, base, "base source must be set (requires yaml3 root-mapping origin fix)")
	require.Contains(t, base.File, "pet.yaml")
	require.Equal(t, 3, base.Line, "base source line must point to '- id' in pet.yaml")
}

// CL: changing write-only required response property to optional
func TestResponseWriteOnlyPropertyBecameOptionalCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_became_optional_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_became_optional_revision.yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	s1.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["name"].Value.WriteOnly = true

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameOptionalCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponseWriteOnlyPropertyBecameOptionalId,
		Args:        []any{"data/name", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_became_optional_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}
