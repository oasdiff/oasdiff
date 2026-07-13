package validate

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// TestRuleIDs_MatchKinInventory scans the pinned kin-openapi source for every
// string its typed validation errors embed, derives the rule IDs through the
// real dispatch, and requires the result to equal the registry exactly. A kin
// bump that renames a string or adds a validation fails here, at test time,
// instead of surfacing as a runtime spec-validation-error demotion. Each
// extraction carries a minimum count so scanner rot is also loud.
func TestRuleIDs_MatchKinInventory(t *testing.T) {
	src := kinOpenapi3Source(t)

	extract := func(re string, min int) []string {
		matches := regexp.MustCompile(re).FindAllStringSubmatch(src, -1)
		require.GreaterOrEqualf(t, len(matches), min,
			"scanner found too few sites for %s; kin restructured its constructors — update the scan", re)
		out := make([]string, 0, len(matches))
		for _, m := range matches {
			out = append(out, m[1])
		}
		return out
	}

	var errs []error
	for _, f := range extract(`newRequiredField\("([^"]+)"`, 20) {
		errs = append(errs, &openapi3.RequiredFieldError{Field: f})
	}
	for _, f := range extract(`newFieldVersionMismatch\("([^"]+)"`, 4) {
		errs = append(errs, &openapi3.FieldVersionMismatchError{Field: f, MinVersion: "3.1"})
	}
	for _, f := range extract(`reject\("([^"]+)"\)`, 20) { // schema fields gated to 3.1+
		errs = append(errs, &openapi3.FieldVersionMismatchError{Field: f, MinVersion: "3.1"})
	}
	for _, f := range extract(`errFieldFor32Plus\("([^"]+)"`, 1) {
		errs = append(errs, &openapi3.FieldVersionMismatchError{Field: f, MinVersion: "3.2"})
	}
	for _, k := range extract(`newSchemaValueError\("([^"]+)"`, 2) {
		errs = append(errs, &openapi3.SchemaValueError{ValueKind: k})
	}
	for _, m := range regexp.MustCompile(`newMutuallyExclusiveFields\("([^"]+)", "([^"]+)"`).FindAllStringSubmatch(src, -1) {
		errs = append(errs, &openapi3.MutuallyExclusiveFieldsError{Field1: m[1], Field2: m[2]})
	}
	for _, f := range extract(`newForbiddenField\("([^"]+)"`, 5) {
		errs = append(errs, &openapi3.ForbiddenFieldError{Field: f})
	}
	for _, list := range extract(`newEitherFieldRequired\(\[\]string\{([^}]+)\}`, 2) {
		errs = append(errs, &openapi3.EitherFieldRequiredError{Fields: quotedStrings(list)})
	}
	for _, list := range extract(`newExactlyOneField\(\[\]string\{([^}]+)\}`, 1) {
		errs = append(errs, &openapi3.ExactlyOneFieldError{Fields: quotedStrings(list)})
	}
	for _, f := range extract(`newSchemaBothForms\("([^"]+)"`, 3) {
		errs = append(errs, &openapi3.SchemaBothFormsExclusive{Field: f})
	}
	for _, s := range extract(`newSingleEntryContent\("([^"]+)"`, 2) {
		errs = append(errs, &openapi3.SingleEntryContentError{Subject: s})
	}

	// fixed-literal arms: anchored to kin types, so a kin type change fails the
	// build; listed here so the registry comparison is exact in both directions
	errs = append(errs,
		&openapi3.PathParametersError{},
		&openapi3.ServerURLTemplateError{},
		&openapi3.WebhookNilError{},
		&openapi3.PathParameterRequiredError{},
		&openapi3.DuplicateOperationIDError{},
		&openapi3.ExtraSiblingFieldsError{},
		&openapi3.SchemaTypeError{},
		&openapi3.InvalidParameterInError{},
		&openapi3.SchemaPatternRegexError{},
		&openapi3.InvalidSecuritySchemeTypeError{},
		&openapi3.InvalidHTTPSchemeError{},
		&openapi3.UnresolvedRefError{},
		&openapi3.APIKeyInInvalidError{},
		&openapi3.PathMustStartWithSlashError{},
		&openapi3.ConflictingPathsError{},
		&openapi3.DuplicateParameterError{},
		&openapi3.InvalidSerializationMethodError{},
	)

	expected := map[string]struct{}{DuplicateEnumValueID: {}, unknownValidationID: {}}
	for _, e := range errs {
		id := ruleIDForKinError(e)
		require.NotEqualf(t, unknownValidationID, id, "scanned kin error %T has no dispatch arm", e)
		expected[id] = struct{}{}
	}
	expectedSorted := make([]string, 0, len(expected))
	for id := range expected {
		expectedSorted = append(expectedSorted, id)
	}
	require.ElementsMatch(t, ruleIDs, expectedSorted,
		"registry and kin-derived inventory diverge: a missing element here means kin added or renamed a validation (add/alias the id in rule_ids.go); an extra element means a registry entry is no longer derivable (remove it or keep it as a documented alias)")
}

// kinOpenapi3Source returns the concatenated non-test Go source of the pinned
// kin-openapi openapi3 package.
func kinOpenapi3Source(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/getkin/kin-openapi").Output()
	require.NoError(t, err, "locate the pinned kin-openapi module")
	dir := filepath.Join(strings.TrimSpace(string(out)), "openapi3")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	var b strings.Builder
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		require.NoError(t, err)
		b.Write(data)
		b.WriteByte('\n')
	}
	return b.String()
}

// quotedStrings extracts the quoted elements of a Go string-slice literal body.
func quotedStrings(list string) []string {
	var out []string
	for _, m := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(list, -1) {
		out = append(out, m[1])
	}
	return out
}
