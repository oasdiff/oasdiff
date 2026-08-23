package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/oasdiff/oasdiff/validate"
	"github.com/stretchr/testify/require"
)

// A location is only known to be *right* when a human has looked at a fixture
// and written down the line it should report. Nothing can derive that. What
// these tests can do is make sure the judgement happens: every rule either has
// a fixture pinning its location, or is listed below as one nobody has got to.
//
// locationUnpinned is a ledger, not an approval. An entry means the rule's
// location is unverified, not that it is exempt. Shrink it by adding a fixture
// under data/validate that triggers the rule and asserting the expected line,
// the way Test_ValidateCmd_FieldVersionMismatch_LicenseIdentifier does.
//
// A kin release that adds a validation code walks the existing chain first
// (TestRuleIDs_MatchKinCatalog, then TestRuleDescriptions_CoverEveryRuleID) and
// then fails here, unpinned and unlisted, so whoever adds the code has to
// decide whether its location deserves a fixture or an entry below.
var locationUnpinned = []string{
	"additional-operations-field-for-3-2-plus",
	"additional-properties-both-forms-exclusive",
	"ambiguous-parameter-serialization",
	"anchor-field-for-3-1-plus",
	"authorization-url-forbidden",
	"bearer-format-forbidden",
	"boolean-schema-for-3-1-plus",
	"boolean-schema-with-other-keywords",
	"comment-field-for-3-1-plus",
	"const-not-in-enum",
	"contains-field-for-3-1-plus",
	"content-encoding-field-for-3-1-plus",
	"content-media-type-field-for-3-1-plus",
	"content-or-schema-exactly-one",
	"content-schema-field-for-3-1-plus",
	"default-required",
	"default-violates-schema",
	"defs-field-for-3-1-plus",
	"dependent-required-field-for-3-1-plus",
	"dependent-schemas-field-for-3-1-plus",
	"duplicate-enum-value",
	"dynamic-anchor-field-for-3-1-plus",
	"dynamic-ref-field-for-3-1-plus",
	"else-field-for-3-1-plus",
	"enum-empty",
	"example-examples-mutually-exclusive",
	"examples-field-for-3-1-plus",
	"external-docs-url-required",
	"flows-forbidden",
	"flows-required",
	"header-content-single-entry",
	"id-field-for-3-1-plus",
	"if-field-for-3-1-plus",
	"in-forbidden",
	"info-required",
	"info-title-required",
	"item-schema-field-for-3-2-plus",
	"json-schema-dialect-required",
	"jsonschemadialect-field-for-3-1-plus",
	"license-name-required",
	"max-contains-field-for-3-1-plus",
	"min-contains-exceeds-max-contains",
	"min-contains-field-for-3-1-plus",
	"min-items-exceeds-max-items",
	"min-length-exceeds-max-length",
	"min-properties-exceeds-max-properties",
	"minimum-exceeds-maximum",
	"multiple-of-not-positive",
	"name-forbidden",
	"oauth-flow-authorization-url-required",
	"oauth-flow-scopes-required",
	"oauth-flow-token-url-required",
	"openid-connect-url-required",
	"operation-id-operation-ref-mutually-exclusive",
	"operation-id-or-operation-ref-required",
	"parameter-content-single-entry",
	"parameter-name-required",
	"path-parameters-mismatch",
	"paths-required",
	"pattern-properties-field-for-3-1-plus",
	"prefix-items-field-for-3-1-plus",
	"property-names-field-for-3-1-plus",
	"query-field-for-3-2-plus",
	"read-only-write-only-mutually-exclusive",
	"request-body-content-required",
	"required-with-default",
	"response-description-required",
	"responses-required",
	"schema-field-for-3-1-plus",
	"schema-items-required",
	"security-scheme-name-required",
	"serialization-method-invalid",
	"server-url-required",
	"spec-validation-error",
	"subschemas-empty",
	"summary-field-for-3-1-plus",
	"then-field-for-3-1-plus",
	"token-url-forbidden",
	"type-format-mismatch",
	"unevaluated-items-both-forms-exclusive",
	"unevaluated-items-field-for-3-1-plus",
	"unevaluated-properties-both-forms-exclusive",
	"unevaluated-properties-field-for-3-1-plus",
	"unresolved-ref",
	"url-identifier-mutually-exclusive",
	"value-external-value-mutually-exclusive",
	"value-or-external-value-required",
	"webhook-nil",
}

// rulesPinnedByFixtures returns the rules that the data/validate corpus
// triggers *with* a source location.
func rulesPinnedByFixtures(t *testing.T) []string {
	t.Helper()

	fixtures, err := filepath.Glob("../data/validate/*.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, fixtures, "no validate fixtures found")

	pinned := map[string]bool{}
	for _, fixture := range fixtures {
		var stdout bytes.Buffer
		internal.Run(cmdToArgs("oasdiff validate -f json --fail-on INFO "+fixture), &stdout, io.Discard)

		var findings []struct {
			Id     string `json:"id"`
			Source struct {
				Line int `json:"line"`
			} `json:"source"`
		}
		require.NoError(t, json.Unmarshal(stdout.Bytes(), &findings), "fixture %s", fixture)

		for _, finding := range findings {
			require.NotZerof(t, finding.Source.Line,
				"%s reports %s with no source location; every finding a fixture produces must have one",
				fixture, finding.Id)
			pinned[finding.Id] = true
		}
	}

	ids := make([]string, 0, len(pinned))
	for id := range pinned {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// Every rule is either pinned by a fixture or on the ledger. A new rule is
// neither, so it fails here until someone decides which it is.
func Test_ValidateCmd_EveryRuleLocationIsPinnedOrLedgered(t *testing.T) {
	pinned := rulesPinnedByFixtures(t)

	for _, id := range validate.RuleIDs() {
		if slices.Contains(pinned, id) || slices.Contains(locationUnpinned, id) {
			continue
		}
		t.Errorf("rule %q has no fixture pinning its source location and is not in locationUnpinned: "+
			"add a data/validate fixture asserting the line it should report, or add it to the ledger", id)
	}
}

// The ledger only shrinks. Once a fixture pins a rule, its entry has to go, so
// the list can't quietly keep claiming coverage is missing where it isn't.
func Test_ValidateCmd_LocationLedgerHasNoStaleEntries(t *testing.T) {
	pinned := rulesPinnedByFixtures(t)

	require.True(t, slices.IsSorted(locationUnpinned), "keep locationUnpinned sorted")
	for _, id := range locationUnpinned {
		require.Containsf(t, validate.RuleIDs(), id, "locationUnpinned lists %q, which is not a rule", id)
		require.NotContainsf(t, pinned, id,
			"%q is pinned by a fixture now; remove it from locationUnpinned", id)
	}
}
