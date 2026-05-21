package checker_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestXOfInlineEnumRefactorToRefDoesNotReportRemoved(t *testing.T) {
	tests := []struct {
		name      string
		check     checker.BackwardCompatibilityCheck
		removedID string
		base      string
		revision  string
	}{
		{
			name:      "request body anyOf",
			check:     checker.RequestPropertyAnyOfUpdatedCheck,
			removedID: checker.RequestBodyAnyOfRemovedId,
			base:      xOfBodySpec("request", "anyOf", false),
			revision:  xOfBodySpec("request", "anyOf", true),
		},
		{
			name:      "request property anyOf",
			check:     checker.RequestPropertyAnyOfUpdatedCheck,
			removedID: checker.RequestPropertyAnyOfRemovedId,
			base:      xOfPropertySpec("request", "anyOf", false),
			revision:  xOfPropertySpec("request", "anyOf", true),
		},
		{
			name:      "response body anyOf",
			check:     checker.ResponsePropertyAnyOfUpdatedCheck,
			removedID: checker.ResponseBodyAnyOfRemovedId,
			base:      xOfBodySpec("response", "anyOf", false),
			revision:  xOfBodySpec("response", "anyOf", true),
		},
		{
			name:      "response property anyOf",
			check:     checker.ResponsePropertyAnyOfUpdatedCheck,
			removedID: checker.ResponsePropertyAnyOfRemovedId,
			base:      xOfPropertySpec("response", "anyOf", false),
			revision:  xOfPropertySpec("response", "anyOf", true),
		},
		{
			name:      "request body oneOf",
			check:     checker.RequestPropertyOneOfUpdatedCheck,
			removedID: checker.RequestBodyOneOfRemovedId,
			base:      xOfBodySpec("request", "oneOf", false),
			revision:  xOfBodySpec("request", "oneOf", true),
		},
		{
			name:      "request property oneOf",
			check:     checker.RequestPropertyOneOfUpdatedCheck,
			removedID: checker.RequestPropertyOneOfRemovedId,
			base:      xOfPropertySpec("request", "oneOf", false),
			revision:  xOfPropertySpec("request", "oneOf", true),
		},
		{
			name:      "response body oneOf",
			check:     checker.ResponsePropertyOneOfUpdated,
			removedID: checker.ResponseBodyOneOfRemovedId,
			base:      xOfBodySpec("response", "oneOf", false),
			revision:  xOfBodySpec("response", "oneOf", true),
		},
		{
			name:      "response property oneOf",
			check:     checker.ResponsePropertyOneOfUpdated,
			removedID: checker.ResponsePropertyOneOfRemovedId,
			base:      xOfPropertySpec("response", "oneOf", false),
			revision:  xOfPropertySpec("response", "oneOf", true),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base, revision := loadSpecPairFromStrings(t, test.base, test.revision)
			diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
			require.NoError(t, err)

			changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(test.check), diffReport, operationsSources, checker.INFO)
			require.False(t, containsId(changes, test.removedID))
		})
	}
}

func TestXOfComponentRenameRefactorStillReportsRemoved(t *testing.T) {
	tests := []struct {
		name      string
		keyword   string
		check     checker.BackwardCompatibilityCheck
		removedID string
	}{
		{
			name:      "anyOf",
			keyword:   "anyOf",
			check:     checker.RequestPropertyAnyOfUpdatedCheck,
			removedID: checker.RequestPropertyAnyOfRemovedId,
		},
		{
			name:      "oneOf",
			keyword:   "oneOf",
			check:     checker.RequestPropertyOneOfUpdatedCheck,
			removedID: checker.RequestPropertyOneOfRemovedId,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base, revision := loadSpecPairFromStrings(
				t,
				xOfComponentRenameSpec(test.keyword, "UserRoleV1"),
				xOfComponentRenameSpec(test.keyword, "UserRole"),
			)
			diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
			require.NoError(t, err)

			changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(test.check), diffReport, operationsSources, checker.INFO)
			require.True(t, containsId(changes, test.removedID))
		})
	}
}

func loadSpecPairFromStrings(t *testing.T, base, revision string) (baseSpec, revisionSpec *load.SpecInfo) {
	t.Helper()

	dir := t.TempDir()
	baseFile := filepath.Join(dir, "base.yaml")
	revisionFile := filepath.Join(dir, "revision.yaml")
	require.NoError(t, os.WriteFile(baseFile, []byte(base), 0644))
	require.NoError(t, os.WriteFile(revisionFile, []byte(revision), 0644))

	baseSpec, err := open(baseFile)
	require.NoError(t, err)
	revisionSpec, err = open(revisionFile)
	require.NoError(t, err)

	return baseSpec, revisionSpec
}

func xOfBodySpec(direction, keyword string, withRef bool) string {
	return fmt.Sprintf(`openapi: 3.1.0
info:
  title: Repro
  version: 1.0.0
paths:
  /users:
    %s
components:
  schemas:
%s
`, operationSpec(direction, keyword, false), userRoleComponents("Role", xOfSchema(keyword, withRef), withRef))
}

func xOfPropertySpec(direction, keyword string, withRef bool) string {
	schema := fmt.Sprintf(`      type: object
      properties:
        role:
%s`, indent(xOfSchema(keyword, withRef), 10))

	return fmt.Sprintf(`openapi: 3.1.0
info:
  title: Repro
  version: 1.0.0
paths:
  /users:
    %s
components:
  schemas:
%s
`, operationSpec(direction, keyword, true), userRoleComponents("Payload", schema, withRef))
}

func xOfComponentRenameSpec(keyword, component string) string {
	return fmt.Sprintf(`openapi: 3.1.0
info:
  title: Repro
  version: 1.0.0
paths:
  /users:
    put:
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Payload"
      responses:
        "200":
          description: OK
components:
  schemas:
    Payload:
      type: object
      properties:
        role:
          %s:
            - $ref: "#/components/schemas/%s"
            - type: "null"
    %s:
      type: string
      enum: [user, superadmin]
`, keyword, component, component)
}

func operationSpec(direction, keyword string, withProperty bool) string {
	schemaRef := "#/components/schemas/Role"
	if withProperty {
		schemaRef = "#/components/schemas/Payload"
	}

	if direction == "request" {
		return fmt.Sprintf(`put:
      requestBody:
        content:
          application/json:
            schema:
              $ref: "%s"
      responses:
        "200":
          description: OK`, schemaRef)
	}

	return fmt.Sprintf(`get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "%s"`, schemaRef)
}

func userRoleComponents(name, schema string, withRef bool) string {
	components := fmt.Sprintf(`    %s:
%s`, name, indent(schema, 6))
	if withRef {
		components += `    UserRole:
      type: string
      enum: [user, superadmin]
      title: UserRole
`
	}

	return components
}

func xOfSchema(keyword string, withRef bool) string {
	if withRef {
		return fmt.Sprintf(`%s:
  - $ref: "#/components/schemas/UserRole"
  - type: "null"
`, keyword)
	}

	return fmt.Sprintf(`%s:
  - type: string
    enum: [user, superadmin]
  - type: "null"
`, keyword)
}

func indent(value string, spaces int) string {
	prefix := fmt.Sprintf("%*s", spaces, "")
	result := ""
	for _, line := range strings.Split(value, "\n") {
		if line == "" {
			continue
		}
		result += prefix + line + "\n"
	}
	return result
}
