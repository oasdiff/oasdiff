package diff

import "github.com/getkin/kin-openapi/openapi3"

// SchemaRefsValidationEquivalent reports whether two resolved schema refs have
// the same validation contract according to oasdiff's schema diff model.
// Annotation-only changes such as title, description, examples, default, and
// comments are ignored; checker-significant metadata such as deprecated is
// treated as a contract change.
//
// Uses a fresh diff state so this predicate does not interact with any
// in-progress diff traversal.
func SchemaRefsValidationEquivalent(config *Config, schemaRef1, schemaRef2 *openapi3.SchemaRef) bool {
	schemaDiff, err := getSchemaDiff(config, newState(), schemaRef1, schemaRef2)
	if err != nil {
		return false
	}

	return !schemaDiffHasValidationChanges(schemaDiff)
}

func schemaDiffHasValidationChanges(schemaDiff *SchemaDiff) bool {
	validationDiff := schemaDiffWithoutAnnotationChanges(schemaDiff)
	return !validationDiff.Empty()
}

func schemaDiffWithoutAnnotationChanges(schemaDiff *SchemaDiff) *SchemaDiff {
	if schemaDiff == nil {
		return nil
	}

	result := *schemaDiff

	result.TitleDiff = nil
	result.DescriptionDiff = nil
	result.DefaultDiff = nil
	result.ExampleDiff = nil
	result.ExternalDocsDiff = nil
	result.ExamplesDiff = nil
	result.CommentDiff = nil

	result.OneOfDiff = subschemasDiffWithoutAnnotationChanges(schemaDiff.OneOfDiff)
	result.AnyOfDiff = subschemasDiffWithoutAnnotationChanges(schemaDiff.AnyOfDiff)
	result.AllOfDiff = subschemasDiffWithoutAnnotationChanges(schemaDiff.AllOfDiff)
	result.NotDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.NotDiff)
	result.ItemsDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.ItemsDiff)
	result.PropertiesDiff = schemasDiffWithoutAnnotationChanges(schemaDiff.PropertiesDiff)
	result.AdditionalPropertiesDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.AdditionalPropertiesDiff)
	result.PrefixItemsDiff = subschemasDiffWithoutAnnotationChanges(schemaDiff.PrefixItemsDiff)
	result.ContainsDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.ContainsDiff)
	result.PatternPropertiesDiff = schemasDiffWithoutAnnotationChanges(schemaDiff.PatternPropertiesDiff)
	result.DependentSchemasDiff = schemasDiffWithoutAnnotationChanges(schemaDiff.DependentSchemasDiff)
	result.PropertyNamesDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.PropertyNamesDiff)
	result.UnevaluatedItemsDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.UnevaluatedItemsDiff)
	result.UnevaluatedPropertiesDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.UnevaluatedPropertiesDiff)
	result.IfDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.IfDiff)
	result.ThenDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.ThenDiff)
	result.ElseDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.ElseDiff)
	result.ContentSchemaDiff = schemaDiffWithoutAnnotationChanges(schemaDiff.ContentSchemaDiff)
	result.DefsDiff = schemasDiffWithoutAnnotationChanges(schemaDiff.DefsDiff)

	if result.Empty() {
		return nil
	}

	return &result
}

func subschemasDiffWithoutAnnotationChanges(subschemasDiff *SubschemasDiff) *SubschemasDiff {
	if subschemasDiff == nil {
		return nil
	}

	result := *subschemasDiff
	result.Modified = ModifiedSubschemas{}

	for _, modified := range subschemasDiff.Modified {
		if modified == nil {
			continue
		}

		schemaDiff := schemaDiffWithoutAnnotationChanges(modified.Diff)
		if schemaDiff == nil {
			continue
		}

		modifiedCopy := *modified
		modifiedCopy.Diff = schemaDiff
		result.Modified = append(result.Modified, &modifiedCopy)
	}

	if result.Empty() {
		return nil
	}

	return &result
}

func schemasDiffWithoutAnnotationChanges(schemasDiff *SchemasDiff) *SchemasDiff {
	if schemasDiff == nil {
		return nil
	}

	result := *schemasDiff
	result.Modified = ModifiedSchemasMap{}

	for name, schemaDiff := range schemasDiff.Modified {
		validationDiff := schemaDiffWithoutAnnotationChanges(schemaDiff)
		if validationDiff == nil {
			continue
		}
		result.Modified[name] = validationDiff
	}

	if result.Empty() {
		return nil
	}

	return &result
}
