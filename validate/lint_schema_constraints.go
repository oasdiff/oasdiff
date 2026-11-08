package validate

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// Schema-constraint coherence lints: a single schema whose own keywords
// contradict each other or violate a JSON Schema MUST. The parser accepts all
// of these (they are well-formed), but no instance can ever satisfy them, so
// they are authoring mistakes rather than choices. See oasdiff/oasdiff#927.
//
// Severity follows the spec text. A literal MUST violation is an error; a
// combination the spec does not forbid outright but that no value can satisfy
// (or a SHOULD) is a warning.
const (
	// SubschemasEmptyID: allOf / anyOf / oneOf MUST be a non-empty array.
	SubschemasEmptyID = "subschemas-empty"

	// The remaining rules describe unsatisfiable bounds: min above max means no
	// instance can validate.
	MinimumExceedsMaximumID         = "minimum-exceeds-maximum"
	MinLengthExceedsMaxLengthID     = "min-length-exceeds-max-length"
	MinItemsExceedsMaxItemsID       = "min-items-exceeds-max-items"
	MinPropertiesExceedsMaxPropsID  = "min-properties-exceeds-max-properties"
	MinContainsExceedsMaxContainsID = "min-contains-exceeds-max-contains"
	// EnumEmptyID: JSON Schema says an "enum" array SHOULD have at least one
	// element; an empty one accepts nothing.
	EnumEmptyID = "enum-empty"
	// ConstNotInEnumID: "const" and "enum" together restrict to their
	// intersection, so a const outside the enum accepts nothing.
	ConstNotInEnumID = "const-not-in-enum"
	// RequiredPropertyForbiddenID: "additionalProperties: false" allows only the
	// properties the schema declares, so requiring one it does not declare
	// accepts nothing.
	RequiredPropertyForbiddenID = "required-property-forbidden"
)

// lintSchemaConstraints reports every schema whose constraints contradict each
// other. It uses kin-openapi's WalkSchemas so traversal and $ref-cycle guarding
// are owned upstream and each finding's Section is the schema's exact RFC 6901
// JSON Pointer, matching the other native lints.
func lintSchemaConstraints(spec *openapi3.T, source string) formatters.Findings {
	var findings formatters.Findings
	// The callback never errors, so WalkSchemas never returns one.
	_ = spec.WalkSchemas(func(jsonPointer string, schema *openapi3.SchemaRef) error {
		s := schema.Value // WalkSchemas guarantees schema and schema.Value are non-nil
		findings = append(findings, schemaConstraintFindings(s, jsonPointer, source)...)
		return nil
	})
	return findings
}

func schemaConstraintFindings(s *openapi3.Schema, section, source string) formatters.Findings {
	var findings formatters.Findings
	// the level comes from RuleLevel, so a finding and `oasdiff checks validate`
	// cannot disagree about the same rule
	add := func(id string, field, text string, args ...any) {
		findings = append(findings, newSchemaConstraintFinding(s, id, RuleLevel(id), field, text, section, source, args))
	}

	// An explicitly empty allOf / anyOf / oneOf. A nil one is simply absent, so
	// the length check alone would flag every ordinary schema. Ordered slice, not
	// a map, so findings come out in the same order every run.
	for _, sub := range []struct {
		field      string
		subschemas openapi3.SchemaRefs
	}{
		{"allOf", s.AllOf}, {"anyOf", s.AnyOf}, {"oneOf", s.OneOf},
	} {
		field, subschemas := sub.field, sub.subschemas
		if subschemas != nil && len(subschemas) == 0 {
			add(SubschemasEmptyID, field,
				fmt.Sprintf("%s must not be empty", field), field)
		}
	}

	if s.Min != nil && s.Max != nil && *s.Min > *s.Max {
		add(MinimumExceedsMaximumID, "minimum",
			fmt.Sprintf("minimum %v is greater than maximum %v, so no value can validate", *s.Min, *s.Max), *s.Min, *s.Max)
	}
	if s.MaxLength != nil && s.MinLength > *s.MaxLength {
		add(MinLengthExceedsMaxLengthID, "minLength",
			fmt.Sprintf("minLength %d is greater than maxLength %d, so no value can validate", s.MinLength, *s.MaxLength), s.MinLength, *s.MaxLength)
	}
	if s.MaxItems != nil && s.MinItems > *s.MaxItems {
		add(MinItemsExceedsMaxItemsID, "minItems",
			fmt.Sprintf("minItems %d is greater than maxItems %d, so no array can validate", s.MinItems, *s.MaxItems), s.MinItems, *s.MaxItems)
	}
	if s.MaxProps != nil && s.MinProps > *s.MaxProps {
		add(MinPropertiesExceedsMaxPropsID, "minProperties",
			fmt.Sprintf("minProperties %d is greater than maxProperties %d, so no object can validate", s.MinProps, *s.MaxProps), s.MinProps, *s.MaxProps)
	}
	if s.MinContains != nil && s.MaxContains != nil && *s.MinContains > *s.MaxContains {
		add(MinContainsExceedsMaxContainsID, "minContains",
			fmt.Sprintf("minContains %d is greater than maxContains %d, so no array can validate", *s.MinContains, *s.MaxContains), *s.MinContains, *s.MaxContains)
	}

	if s.Enum != nil && len(s.Enum) == 0 {
		add(EnumEmptyID, "enum", "enum is empty, so no value can validate", "enum")
	}
	if s.Const != nil && len(s.Enum) > 0 && !enumContains(s.Enum, s.Const) {
		add(ConstNotInEnumID, "const",
			fmt.Sprintf("const %s is not one of the enum values, so no value can validate", displayEnum(s.Const)), displayEnum(s.Const))
	}

	// additionalProperties as a schema allows an undeclared property that
	// validates against it, and a patternProperties entry may allow one too, so
	// only a plain false with no patterns forbids every undeclared property.
	if s.AdditionalProperties.Has != nil && !*s.AdditionalProperties.Has && len(s.PatternProperties) == 0 {
		for _, name := range s.Required {
			if s.Properties[name] == nil {
				add(RequiredPropertyForbiddenID, "required",
					fmt.Sprintf("required property %s is not declared in properties, and additionalProperties is false, so no object can validate", name), name)
			}
		}
	}

	return findings
}

// enumContains reports whether the enum holds value, comparing by JSON encoding
// so 1 (number) and "1" (string) do not collide. Mirrors duplicateEnumValues.
func enumContains(enum []any, value any) bool {
	target := enumKey(value)
	for _, v := range enum {
		if enumKey(v) == target {
			return true
		}
	}
	return false
}

func newSchemaConstraintFinding(s *openapi3.Schema, id string, level checker.Level, field, text, section, source string, args []any) formatters.Finding {
	line, column := schemaFieldLocation(s, field)
	f := formatters.Finding{
		Id:      id,
		Text:    text,
		Level:   level,
		Section: section,
		Source: formatters.Source{
			File:   source,
			Line:   line,
			Column: column,
		},
	}
	f.Fingerprint = checker.ComputeFingerprint(f.Id, "", section, args)
	return f
}
