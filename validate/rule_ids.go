package validate

import "slices"

// ruleIDs is the registry of every rule ID validate can emit: the fixed, public
// ID surface. Most IDs are the stable codes kin-openapi declares on its
// validation errors (openapi3.ValidationErrorCodes), plus a few oasdiff-native
// lint IDs. TestRuleIDs_MatchKinCatalog pins the registry to that union, so a
// kin code added or renamed on a bump fails the build until the registry is
// updated deliberately. At runtime, a code not in the registry is demoted to
// spec-validation-error (loud, triaged). Sorted; add new IDs in order.
var ruleIDs = []string{
	"additional-properties-both-forms-exclusive",
	"ambiguous-parameter-serialization",
	"anchor-field-for-3-1-plus",
	"authorization-url-forbidden",
	"bearer-format-forbidden",
	"comment-field-for-3-1-plus",
	"conflicting-paths",
	"const-field-for-3-1-plus",
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
	"duplicate-operation-id",
	"duplicate-parameter",
	"duplicate-required-field",
	"duplicate-tag",
	"dynamic-anchor-field-for-3-1-plus",
	"dynamic-ref-field-for-3-1-plus",
	"else-field-for-3-1-plus",
	"example-examples-mutually-exclusive",
	"example-violates-schema",
	"examples-field-for-3-1-plus",
	"external-docs-url-required",
	"extra-sibling-fields",
	"flows-forbidden",
	"flows-required",
	"header-content-single-entry",
	"id-field-for-3-1-plus",
	"identifier-field-for-3-1-plus",
	"if-field-for-3-1-plus",
	"in-forbidden",
	"info-required",
	"info-title-required",
	"info-version-required",
	"item-schema-field-for-3-2-plus",
	"json-schema-dialect-required",
	"jsonschemadialect-field-for-3-1-plus",
	"license-name-required",
	"max-contains-field-for-3-1-plus",
	"min-contains-field-for-3-1-plus",
	"name-forbidden",
	"oauth-flow-authorization-url-required",
	"oauth-flow-scopes-required",
	"oauth-flow-token-url-required",
	"openapi-required",
	"openid-connect-url-required",
	"operation-id-operation-ref-mutually-exclusive",
	"operation-id-or-operation-ref-required",
	"operation-responses-required",
	"parameter-content-single-entry",
	"parameter-in-invalid",
	"parameter-name-required",
	"path-must-start-with-slash",
	"path-parameter-required",
	"path-parameters-mismatch",
	"paths-required",
	"pattern-properties-field-for-3-1-plus",
	"prefix-items-field-for-3-1-plus",
	"property-names-field-for-3-1-plus",
	"read-only-write-only-mutually-exclusive",
	"request-body-content-required",
	"required-with-default",
	"response-description-required",
	"responses-required",
	"schema-field-for-3-1-plus",
	"schema-items-required",
	"schema-pattern-regex-invalid",
	"schema-type-unsupported",
	"security-scheme-apikey-in-invalid",
	"security-scheme-http-scheme-invalid",
	"security-scheme-name-required",
	"security-scheme-type-invalid",
	"serialization-method-invalid",
	"server-url-required",
	"server-url-template-invalid",
	"spec-validation-error",
	"summary-field-for-3-1-plus",
	"then-field-for-3-1-plus",
	"token-url-forbidden",
	"unevaluated-items-both-forms-exclusive",
	"unevaluated-items-field-for-3-1-plus",
	"unevaluated-properties-both-forms-exclusive",
	"unevaluated-properties-field-for-3-1-plus",
	"unresolved-ref",
	"url-identifier-mutually-exclusive",
	"value-external-value-mutually-exclusive",
	"value-or-external-value-required",
	"webhook-nil",
	"webhooks-field-for-3-1-plus",
}

// nativeRuleIDs are the ids validate emits that do not come from kin: the
// oasdiff-native lints and the unknown-error fallback. Kept apart so
// TestRuleIDs_MatchKinCatalog can assert ruleIDs == kin's codes ∪ these.
var nativeRuleIDs = []string{
	AmbiguousParameterSerializationID,
	DuplicateEnumValueID,
	RequiredWithDefaultID,
	unknownValidationID,
}

// RuleIDs returns every rule ID validate can emit, sorted.
func RuleIDs() []string {
	return slices.Clone(ruleIDs)
}

// knownRuleID returns id if it is in the registry, else spec-validation-error.
func knownRuleID(id string) string {
	if _, found := slices.BinarySearch(ruleIDs, id); found {
		return id
	}
	return unknownValidationID
}
