package diff

import "github.com/getkin/kin-openapi/openapi3"

// SchemaRefsValidationEquivalent reports whether two resolved schema refs have
// the same validation contract according to oasdiff's schema diff model.
// Annotation-only changes such as title, description, examples, default, and
// comments are ignored. Checker-significant metadata such as deprecated is
// treated as a contract change.
func SchemaRefsValidationEquivalent(schemaRef1, schemaRef2 *openapi3.SchemaRef) bool {
	schemaDiff, err := getSchemaDiff(NewConfig(), newState(), schemaRef1, schemaRef2)
	if err != nil {
		return false
	}

	return !schemaDiffHasValidationChanges(schemaDiff)
}

func schemaDiffHasValidationChanges(schemaDiff *SchemaDiff) bool {
	if schemaDiff == nil {
		return false
	}

	if schemaDiff.SchemaAdded ||
		schemaDiff.SchemaDeleted ||
		schemaDiff.CircularRefDiff ||
		!schemaDiff.ExtensionsDiff.Empty() ||
		subschemasDiffHasValidationChanges(schemaDiff.OneOfDiff) ||
		subschemasDiffHasValidationChanges(schemaDiff.AnyOfDiff) ||
		subschemasDiffHasValidationChanges(schemaDiff.AllOfDiff) ||
		schemaDiffHasValidationChanges(schemaDiff.NotDiff) ||
		!schemaDiff.TypeDiff.Empty() ||
		!schemaDiff.ListOfTypesDiff.Empty() ||
		!schemaDiff.FormatDiff.Empty() ||
		!schemaDiff.EnumDiff.Empty() ||
		!schemaDiff.AdditionalPropertiesAllowedDiff.Empty() ||
		!schemaDiff.UniqueItemsDiff.Empty() ||
		!schemaDiff.ExclusiveMinDiff.Empty() ||
		!schemaDiff.ExclusiveMaxDiff.Empty() ||
		!schemaDiff.NullableDiff.Empty() ||
		!schemaDiff.ReadOnlyDiff.Empty() ||
		!schemaDiff.WriteOnlyDiff.Empty() ||
		!schemaDiff.AllowEmptyValueDiff.Empty() ||
		!schemaDiff.XMLDiff.Empty() ||
		!schemaDiff.DeprecatedDiff.Empty() ||
		!schemaDiff.MinDiff.Empty() ||
		!schemaDiff.MaxDiff.Empty() ||
		!schemaDiff.MultipleOfDiff.Empty() ||
		!schemaDiff.MinLengthDiff.Empty() ||
		!schemaDiff.MaxLengthDiff.Empty() ||
		!schemaDiff.PatternDiff.Empty() ||
		!schemaDiff.MinItemsDiff.Empty() ||
		!schemaDiff.MaxItemsDiff.Empty() ||
		schemaDiffHasValidationChanges(schemaDiff.ItemsDiff) ||
		!schemaDiff.RequiredDiff.Empty() ||
		schemasDiffHasValidationChanges(schemaDiff.PropertiesDiff) ||
		!schemaDiff.MinPropsDiff.Empty() ||
		!schemaDiff.MaxPropsDiff.Empty() ||
		schemaDiffHasValidationChanges(schemaDiff.AdditionalPropertiesDiff) ||
		!schemaDiff.DiscriminatorDiff.Empty() ||
		!schemaDiff.ConstDiff.Empty() ||
		subschemasDiffHasValidationChanges(schemaDiff.PrefixItemsDiff) ||
		schemaDiffHasValidationChanges(schemaDiff.ContainsDiff) ||
		!schemaDiff.MinContainsDiff.Empty() ||
		!schemaDiff.MaxContainsDiff.Empty() ||
		schemasDiffHasValidationChanges(schemaDiff.PatternPropertiesDiff) ||
		schemasDiffHasValidationChanges(schemaDiff.DependentSchemasDiff) ||
		schemaDiffHasValidationChanges(schemaDiff.PropertyNamesDiff) ||
		!schemaDiff.UnevaluatedItemsAllowedDiff.Empty() ||
		schemaDiffHasValidationChanges(schemaDiff.UnevaluatedItemsDiff) ||
		!schemaDiff.UnevaluatedPropertiesAllowedDiff.Empty() ||
		schemaDiffHasValidationChanges(schemaDiff.UnevaluatedPropertiesDiff) ||
		schemaDiffHasValidationChanges(schemaDiff.IfDiff) ||
		schemaDiffHasValidationChanges(schemaDiff.ThenDiff) ||
		schemaDiffHasValidationChanges(schemaDiff.ElseDiff) ||
		!schemaDiff.DependentRequiredDiff.Empty() ||
		!schemaDiff.SchemaIDDiff.Empty() ||
		!schemaDiff.AnchorDiff.Empty() ||
		!schemaDiff.DynamicRefDiff.Empty() ||
		!schemaDiff.DynamicAnchorDiff.Empty() ||
		!schemaDiff.ContentMediaTypeDiff.Empty() ||
		!schemaDiff.ContentEncodingDiff.Empty() ||
		schemaDiffHasValidationChanges(schemaDiff.ContentSchemaDiff) ||
		schemasDiffHasValidationChanges(schemaDiff.DefsDiff) ||
		!schemaDiff.SchemaDialectDiff.Empty() {
		return true
	}

	return false
}

func subschemasDiffHasValidationChanges(subschemasDiff *SubschemasDiff) bool {
	if subschemasDiff == nil {
		return false
	}

	if len(subschemasDiff.Added) > 0 || len(subschemasDiff.Deleted) > 0 {
		return true
	}

	for _, modified := range subschemasDiff.Modified {
		if modified == nil {
			continue
		}
		if schemaDiffHasValidationChanges(modified.Diff) {
			return true
		}
	}

	return false
}

func schemasDiffHasValidationChanges(schemasDiff *SchemasDiff) bool {
	if schemasDiff == nil {
		return false
	}

	if len(schemasDiff.Added) > 0 || len(schemasDiff.Deleted) > 0 {
		return true
	}

	for _, schemaDiff := range schemasDiff.Modified {
		if schemaDiffHasValidationChanges(schemaDiff) {
			return true
		}
	}

	return false
}
