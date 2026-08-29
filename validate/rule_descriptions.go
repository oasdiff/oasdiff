package validate

import (
	"regexp"
	"strings"
)

// Descriptions for the rules `oasdiff validate` can report, used by
// `oasdiff checks validate` to say what each ID means.
//
// Plain English, not localized: unlike the changelog rules, a validate finding
// takes its message from the parser at runtime, so there is no localized
// message file to hang these off. They describe the rule, not the specific
// violation; the finding's own text names the offending field.
//
// Style follows the changelog rule descriptions: short, lowercase, no trailing
// period.
//
// Not every ID is listed. The version-gate rules (`<field>-field-for-3-1-plus`
// and friends) are described by versionGateDescription, which derives the text
// from the ID, so a new one added upstream is described without a change here.
var ruleDescriptions = map[string]string{
	// Document structure
	"openapi-required":             "openapi version field is missing",
	"info-required":                "info object is missing",
	"info-title-required":          "info title is missing",
	"info-version-required":        "info version is missing",
	"license-name-required":        "license name is missing",
	"paths-required":               "paths object is missing or not an object",
	"json-schema-dialect-required": "jsonSchemaDialect value is missing",
	"external-docs-url-required":   "external documentation url is missing",
	"unresolved-ref":               "a $ref could not be resolved",
	"extra-sibling-fields":         "fields alongside a $ref are ignored",
	"spec-validation-error":        "a spec violation with no more specific rule",

	// Paths and operations
	"path-must-start-with-slash":         "path key does not start with a slash",
	"conflicting-paths":                  "two paths differ only by parameter name and cannot be told apart",
	"path-parameters-mismatch":           "path template and declared path parameters disagree",
	"path-parameter-required":            "a path parameter is not marked required",
	"duplicate-operation-id":             "the same operationId is used by more than one operation",
	"operation-responses-required":       "operation has no responses object",
	"boolean-schema-for-3-1-plus":        "a schema is written as `true` or `false`, which is only valid in OpenAPI 3.1 or later",
	"boolean-schema-with-other-keywords": "a boolean schema carries other keywords, which would be lost when it is written back",
	"responses-required":                 "responses object is missing",
	"response-description-required":      "response has no description",
	"request-body-content-required":      "request body has no content",
	// additionalOperations is the OpenAPI 3.2 map of custom http methods. The
	// version gate on the field itself is described by versionGateDescription.
	"additional-operations-duplicate-method": "an additionalOperations key names a method that already has its own path item field",
	"additional-operations-invalid-method":   "an additionalOperations key is not a valid http method name",

	// Parameters and headers
	"parameter-name-required":        "parameter name is missing",
	"parameter-in-invalid":           "parameter location is not query, header, path or cookie",
	"duplicate-parameter":            "the same parameter is declared more than once",
	"parameter-content-single-entry": "parameter content must hold exactly one media type",
	"header-content-single-entry":    "header content must hold exactly one media type",
	"content-or-schema-exactly-one":  "parameter or header must use either content or schema, not both",
	"name-forbidden":                 "header must not declare a name field",
	"in-forbidden":                   "header must not declare an in field",
	"serialization-method-invalid":   "style and explode combination is not valid for this parameter",

	// Schemas
	"schema-type-unsupported":                     "schema type is not a valid JSON Schema type",
	"schema-items-required":                       "array schema is missing items",
	"schema-pattern-regex-invalid":                "pattern is not a valid regular expression",
	"read-only-write-only-mutually-exclusive":     "schema is both readOnly and writeOnly",
	"additional-properties-both-forms-exclusive":  "additionalProperties is given as both a boolean and a schema",
	"unevaluated-items-both-forms-exclusive":      "unevaluatedItems is given as both a boolean and a schema",
	"unevaluated-properties-both-forms-exclusive": "unevaluatedProperties is given as both a boolean and a schema",
	"duplicate-required-field":                    "the same property is listed more than once in required",
	"default-violates-schema":                     "default value does not satisfy its own schema",
	"example-violates-schema":                     "example does not satisfy its schema",
	"default-required":                            "a default value is missing where one is required",

	// Schema constraint coherence (oasdiff-native)
	"multiple-of-not-positive":              "multipleOf is not greater than zero",
	"subschemas-empty":                      "allOf, anyOf or oneOf is empty",
	"minimum-exceeds-maximum":               "minimum is greater than maximum, so no value can validate",
	"min-length-exceeds-max-length":         "minLength is greater than maxLength, so no value can validate",
	"min-items-exceeds-max-items":           "minItems is greater than maxItems, so no array can validate",
	"min-properties-exceeds-max-properties": "minProperties is greater than maxProperties, so no object can validate",
	"min-contains-exceeds-max-contains":     "minContains is greater than maxContains, so no array can validate",
	"enum-empty":                            "enum is empty, so no value can validate",
	"const-not-in-enum":                     "const is not one of the enum values, so no value can validate",
	"duplicate-enum-value":                  "enum lists the same value more than once",
	"type-format-mismatch":                  "format belongs to a different type and is ignored",
	"required-with-default":                 "a required parameter or property also has a default, which is never used",
	"ambiguous-parameter-serialization":     "parameter type mixes a structured type with a scalar, so its serialization is ambiguous",

	// Examples and links
	"example-examples-mutually-exclusive":           "example and examples are both set",
	"value-external-value-mutually-exclusive":       "example sets both value and externalValue",
	"value-or-external-value-required":              "example sets neither value nor externalValue",
	"operation-id-operation-ref-mutually-exclusive": "link sets both operationId and operationRef",
	"operation-id-or-operation-ref-required":        "link sets neither operationId nor operationRef",
	"url-identifier-mutually-exclusive":             "license sets both url and identifier",

	// Security
	"security-scheme-type-invalid":          "security scheme type is not valid",
	"security-scheme-name-required":         "api key security scheme is missing its name",
	"security-scheme-apikey-in-invalid":     "api key location is not query, header or cookie",
	"security-scheme-http-scheme-invalid":   "http security scheme is missing its scheme",
	"openid-connect-url-required":           "openIdConnect security scheme is missing its url",
	"bearer-format-forbidden":               "bearerFormat is set on a non-bearer security scheme",
	"flows-required":                        "oauth2 security scheme is missing its flows",
	"flows-forbidden":                       "flows is set on a non-oauth2 security scheme",
	"oauth-flow-authorization-url-required": "oauth flow is missing its authorization url",
	"oauth-flow-token-url-required":         "oauth flow is missing its token url",
	"oauth-flow-scopes-required":            "oauth flow is missing its scopes",
	"authorization-url-forbidden":           "authorization url is set on a flow that does not use one",
	"token-url-forbidden":                   "token url is set on a flow that does not use one",

	// Servers, tags, webhooks
	"server-url-required":         "server url is missing",
	"server-url-template-invalid": "server url template is malformed or has an undeclared variable",
	"duplicate-tag":               "the same tag name is declared more than once",
	"webhook-nil":                 "webhooks holds an empty path item",
}

// versionGateRe matches the version-gate rule IDs kin emits for a field that
// only exists in a later OpenAPI version, e.g. "const-field-for-3-1-plus".
var versionGateRe = regexp.MustCompile(`^[a-z0-9-]+-field-for-(\d+)-(\d+)-plus$`)

// versionGateDescription describes a version-gate rule, or returns "" when the
// id is not one. Derived from the ID rather than listed per field so a new
// version-gated field added upstream is described without a change here. The
// finding's own text names the field.
func versionGateDescription(id string) string {
	m := versionGateRe.FindStringSubmatch(id)
	if m == nil {
		return ""
	}
	return "field is only valid in OpenAPI " + m[1] + "." + m[2] + " or later"
}

// RuleDescription returns a short description of what a validate rule reports,
// or "" for an unknown ID.
func RuleDescription(id string) string {
	if d, ok := ruleDescriptions[id]; ok {
		return d
	}
	return versionGateDescription(strings.TrimSpace(id))
}
