package diff

import (
	"errors"

	"github.com/getkin/kin-openapi/openapi3"
)

// SchemaDiff describes the changes between a pair of schema objects: https://swagger.io/specification/#schema-object
type SchemaDiff struct {
	SchemaAdded                     bool                    `json:"schemaAdded,omitempty" yaml:"schemaAdded,omitempty"`
	SchemaDeleted                   bool                    `json:"schemaDeleted,omitempty" yaml:"schemaDeleted,omitempty"`
	CircularRefDiff                 bool                    `json:"circularRef,omitempty" yaml:"circularRef,omitempty"`
	ExtensionsDiff                  *ExtensionsDiff         `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	OneOfDiff                       *SubschemasDiff         `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOfDiff                       *SubschemasDiff         `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	AllOfDiff                       *SubschemasDiff         `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	NotDiff                         *SchemaDiff             `json:"not,omitempty" yaml:"not,omitempty"`
	TypeDiff                        *StringsDiff            `json:"type,omitempty" yaml:"type,omitempty"`
	ListOfTypesDiff                 *ListOfTypesDiff        `json:"listOfTypes,omitempty" yaml:"listOfTypes,omitempty"`
	TitleDiff                       *ValueDiff              `json:"title,omitempty" yaml:"title,omitempty"`
	FormatDiff                      *ValueDiff              `json:"format,omitempty" yaml:"format,omitempty"`
	DescriptionDiff                 *ValueDiff              `json:"description,omitempty" yaml:"description,omitempty"`
	EnumDiff                        *EnumDiff               `json:"enum,omitempty" yaml:"enum,omitempty"`
	DefaultDiff                     *ValueDiff              `json:"default,omitempty" yaml:"default,omitempty"`
	ExampleDiff                     *ValueDiff              `json:"example,omitempty" yaml:"example,omitempty"`
	ExternalDocsDiff                *ExternalDocsDiff       `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	AdditionalPropertiesAllowedDiff *ValueDiff              `json:"additionalPropertiesAllowed,omitempty" yaml:"additionalPropertiesAllowed,omitempty"`
	UniqueItemsDiff                 *ValueDiff              `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	ExclusiveMinDiff                *ValueDiff              `json:"exclusiveMin,omitempty" yaml:"exclusiveMin,omitempty"`
	ExclusiveMaxDiff                *ValueDiff              `json:"exclusiveMax,omitempty" yaml:"exclusiveMax,omitempty"`
	NullableDiff                    *ValueDiff              `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	ReadOnlyDiff                    *ValueDiff              `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnlyDiff                   *ValueDiff              `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	AllowEmptyValueDiff             *ValueDiff              `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	XMLDiff                         *ValueDiff              `json:"XML,omitempty" yaml:"XML,omitempty"`
	DeprecatedDiff                  *ValueDiff              `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	MinDiff                         *ValueDiff              `json:"min,omitempty" yaml:"min,omitempty"`
	MaxDiff                         *ValueDiff              `json:"max,omitempty" yaml:"max,omitempty"`
	MultipleOfDiff                  *ValueDiff              `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	MinLengthDiff                   *ValueDiff              `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLengthDiff                   *ValueDiff              `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	PatternDiff                     *ValueDiff              `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinItemsDiff                    *ValueDiff              `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItemsDiff                    *ValueDiff              `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	ItemsDiff                       *SchemaDiff             `json:"items,omitempty" yaml:"items,omitempty"`
	RequiredDiff                    *RequiredPropertiesDiff `json:"required,omitempty" yaml:"required,omitempty"`
	PropertiesDiff                  *SchemasDiff            `json:"properties,omitempty" yaml:"properties,omitempty"`
	MinPropsDiff                    *ValueDiff              `json:"minProps,omitempty" yaml:"minProps,omitempty"`
	MaxPropsDiff                    *ValueDiff              `json:"maxProps,omitempty" yaml:"maxProps,omitempty"`
	AdditionalPropertiesDiff        *SchemaDiff             `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	DiscriminatorDiff               *DiscriminatorDiff      `json:"discriminatorDiff,omitempty" yaml:"discriminatorDiff,omitempty"`
	// OpenAPI 3.1 / JSON Schema 2020-12 fields
	ConstDiff                 *ValueDiff       `json:"const,omitempty" yaml:"const,omitempty"`
	ExamplesDiff              *ValueDiff       `json:"examples,omitempty" yaml:"examples,omitempty"`
	PrefixItemsDiff           *SubschemasDiff  `json:"prefixItems,omitempty" yaml:"prefixItems,omitempty"`
	ContainsDiff              *SchemaDiff      `json:"contains,omitempty" yaml:"contains,omitempty"`
	MinContainsDiff           *ValueDiff       `json:"minContains,omitempty" yaml:"minContains,omitempty"`
	MaxContainsDiff           *ValueDiff       `json:"maxContains,omitempty" yaml:"maxContains,omitempty"`
	PatternPropertiesDiff     *SchemasDiff     `json:"patternProperties,omitempty" yaml:"patternProperties,omitempty"`
	DependentSchemasDiff      *SchemasDiff     `json:"dependentSchemas,omitempty" yaml:"dependentSchemas,omitempty"`
	PropertyNamesDiff         *SchemaDiff      `json:"propertyNames,omitempty" yaml:"propertyNames,omitempty"`
	UnevaluatedItemsDiff      *SchemaDiff      `json:"unevaluatedItems,omitempty" yaml:"unevaluatedItems,omitempty"`
	UnevaluatedPropertiesDiff *SchemaDiff      `json:"unevaluatedProperties,omitempty" yaml:"unevaluatedProperties,omitempty"`
	IfDiff                    *SchemaDiff      `json:"if,omitempty" yaml:"if,omitempty"`
	ThenDiff                  *SchemaDiff      `json:"then,omitempty" yaml:"then,omitempty"`
	ElseDiff                  *SchemaDiff      `json:"else,omitempty" yaml:"else,omitempty"`
	DependentRequiredDiff     *ValueDiff       `json:"dependentRequired,omitempty" yaml:"dependentRequired,omitempty"`
	SchemaIDDiff              *ValueDiff       `json:"$id,omitempty" yaml:"$id,omitempty"`
	AnchorDiff                *ValueDiff       `json:"$anchor,omitempty" yaml:"$anchor,omitempty"`
	DynamicRefDiff            *ValueDiff       `json:"$dynamicRef,omitempty" yaml:"$dynamicRef,omitempty"`
	DynamicAnchorDiff         *ValueDiff       `json:"$dynamicAnchor,omitempty" yaml:"$dynamicAnchor,omitempty"`
	ContentMediaTypeDiff      *ValueDiff       `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`
	ContentEncodingDiff       *ValueDiff       `json:"contentEncoding,omitempty" yaml:"contentEncoding,omitempty"`
	ContentSchemaDiff         *SchemaDiff      `json:"contentSchema,omitempty" yaml:"contentSchema,omitempty"`
	DefsDiff                  *SchemasDiff     `json:"$defs,omitempty" yaml:"$defs,omitempty"`
	SchemaDialectDiff         *ValueDiff       `json:"$schema,omitempty" yaml:"$schema,omitempty"`
	CommentDiff               *ValueDiff       `json:"$comment,omitempty" yaml:"$comment,omitempty"`
	Base                      *openapi3.Schema `json:"-" yaml:"-"`
	Revision                  *openapi3.Schema `json:"-" yaml:"-"`
}

// Empty indicates whether a change was found in this element
func (diff *SchemaDiff) Empty() bool {
	return diff == nil || *diff == SchemaDiff{Base: diff.Base, Revision: diff.Revision}
}

func getSchemaDiff(config *Config, state *state, schema1, schema2 *openapi3.SchemaRef) (*SchemaDiff, error) {

	if diff, ok := state.cache.get(state.direction, schema1, schema2); ok {
		return diff, nil
	}

	diff, err := getSchemaDiffInternal(config, state, schema1, schema2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		diff = nil
	}

	state.cache.add(state.direction, schema1, schema2, diff)
	return diff, nil
}

func getSchemaDiffInternal(config *Config, state *state, schema1, schema2 *openapi3.SchemaRef) (*SchemaDiff, error) {

	if schema1 == nil && schema2 == nil {
		return nil, nil
	} else if schema1 == nil {
		return &SchemaDiff{SchemaAdded: true}, nil
	} else if schema2 == nil {
		return &SchemaDiff{SchemaDeleted: true}, nil
	}

	value1 := schema1.Value
	if value1 == nil {
		return nil, errors.New("base schema value is nil")
	}

	value2 := schema2.Value
	if value2 == nil {
		return nil, errors.New("revision schema value is nil")
	}

	result := SchemaDiff{
		Base:     value1,
		Revision: value2,
	}

	if status := getCircularRefsDiff(state.visitedSchemasBase, state.visitedSchemasRevision, schema1, schema2); status != circularRefStatusNone {
		switch status {
		case circularRefStatusDiff:
			result.CircularRefDiff = true
			return &result, nil
		case circularRefStatusNoDiff:
			return &result, nil
		}
	}

	// mark visited schema references to avoid infinite loops
	if schema1.Ref != "" {
		state.visitedSchemasBase[schema1.Ref] = struct{}{}
		defer delete(state.visitedSchemasBase, schema1.Ref)
	}

	if schema2.Ref != "" {
		state.visitedSchemasRevision[schema2.Ref] = struct{}{}
		defer delete(state.visitedSchemasRevision, schema2.Ref)
	}

	var err error

	result.ExtensionsDiff, err = getExtensionsDiff(config, value1.Extensions, value2.Extensions)
	if err != nil {
		return nil, err
	}
	result.OneOfDiff, err = getSubschemasDiff(config, state, value1.OneOf, value2.OneOf)
	if err != nil {
		return nil, err
	}
	result.AnyOfDiff, err = getSubschemasDiff(config, state, value1.AnyOf, value2.AnyOf)
	if err != nil {
		return nil, err
	}
	result.AllOfDiff, err = getSubschemasDiff(config, state, value1.AllOf, value2.AllOf)
	if err != nil {
		return nil, err
	}
	result.NotDiff, err = getSchemaDiff(config, state, value1.Not, value2.Not)
	if err != nil {
		return nil, err
	}
	result.TypeDiff = getTypeDiff(value1.Type, value2.Type)
	result.ListOfTypesDiff = getListOfTypesDiff(value1, value2)
	result.TitleDiff = getValueDiffConditional(config.IsExcludeTitle(), value1.Title, value2.Title)
	result.FormatDiff = getValueDiff(value1.Format, value2.Format)
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), value1.Description, value2.Description)
	result.EnumDiff = getEnumDiff(value1.Enum, value2.Enum)
	result.DefaultDiff = getValueDiff(value1.Default, value2.Default)
	result.ExampleDiff = getValueDiffConditional(config.IsExcludeExamples(), value1.Example, value2.Example)
	result.ExternalDocsDiff, err = getExternalDocsDiff(config, value1.ExternalDocs, value2.ExternalDocs)
	if err != nil {
		return nil, err
	}
	result.AdditionalPropertiesAllowedDiff = getBoolRefDiff(value1.AdditionalProperties.Has, value2.AdditionalProperties.Has)
	result.UniqueItemsDiff = getValueDiff(value1.UniqueItems, value2.UniqueItems)
	result.ExclusiveMinDiff = getExclusiveBoundDiff(value1.ExclusiveMin, value2.ExclusiveMin)
	result.ExclusiveMaxDiff = getExclusiveBoundDiff(value1.ExclusiveMax, value2.ExclusiveMax)
	result.NullableDiff = getValueDiff(value1.Nullable, value2.Nullable)
	result.ReadOnlyDiff = getValueDiff(value1.ReadOnly, value2.ReadOnly)
	result.WriteOnlyDiff = getValueDiff(value1.WriteOnly, value2.WriteOnly)
	result.AllowEmptyValueDiff = getValueDiff(value1.AllowEmptyValue, value2.AllowEmptyValue)
	result.XMLDiff = getValueDiff(value1.XML, value2.XML)
	result.DeprecatedDiff = getValueDiff(value1.Deprecated, value2.Deprecated)
	result.MinDiff = getFloat64RefDiff(value1.Min, value2.Min)
	result.MaxDiff = getFloat64RefDiff(value1.Max, value2.Max)
	result.MultipleOfDiff = getFloat64RefDiff(value1.MultipleOf, value2.MultipleOf)
	result.MinLengthDiff = getValueDiff(value1.MinLength, value2.MinLength)
	result.MaxLengthDiff = getUInt64RefDiff(value1.MaxLength, value2.MaxLength)
	result.PatternDiff = getValueDiff(value1.Pattern, value2.Pattern)
	// compiledPattern is derived from pattern -> no need to diff
	result.MinItemsDiff = getValueDiff(value1.MinItems, value2.MinItems)
	result.MaxItemsDiff = getUInt64RefDiff(value1.MaxItems, value2.MaxItems)
	result.ItemsDiff, err = getSchemaDiff(config, state, value1.Items, value2.Items)
	if err != nil {
		return nil, err
	}

	// Object
	result.RequiredDiff = getRequiredPropertiesDiff(value1, value2)
	result.PropertiesDiff, err = getSchemasDiff(config, state, value1.Properties, value2.Properties)
	if err != nil {
		return nil, err
	}

	result.MinPropsDiff = getValueDiff(value1.MinProps, value2.MinProps)
	result.MaxPropsDiff = getUInt64RefDiff(value1.MaxProps, value2.MaxProps)
	result.AdditionalPropertiesDiff, err = getSchemaDiff(config, state, value1.AdditionalProperties.Schema, value2.AdditionalProperties.Schema)
	if err != nil {
		return nil, err
	}

	result.DiscriminatorDiff, err = getDiscriminatorDiff(config, value1.Discriminator, value2.Discriminator)
	if err != nil {
		return nil, err
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 fields
	result.ConstDiff = getValueDiff(value1.Const, value2.Const)
	result.ExamplesDiff = getValueDiff(value1.Examples, value2.Examples)
	result.PrefixItemsDiff, err = getSubschemasDiff(config, state, value1.PrefixItems, value2.PrefixItems)
	if err != nil {
		return nil, err
	}
	result.ContainsDiff, err = getSchemaDiff(config, state, value1.Contains, value2.Contains)
	if err != nil {
		return nil, err
	}
	result.MinContainsDiff = getUInt64RefDiff(value1.MinContains, value2.MinContains)
	result.MaxContainsDiff = getUInt64RefDiff(value1.MaxContains, value2.MaxContains)
	result.PatternPropertiesDiff, err = getSchemasDiff(config, state, value1.PatternProperties, value2.PatternProperties)
	if err != nil {
		return nil, err
	}
	result.DependentSchemasDiff, err = getSchemasDiff(config, state, value1.DependentSchemas, value2.DependentSchemas)
	if err != nil {
		return nil, err
	}
	result.PropertyNamesDiff, err = getSchemaDiff(config, state, value1.PropertyNames, value2.PropertyNames)
	if err != nil {
		return nil, err
	}
	result.UnevaluatedItemsDiff, err = getSchemaDiff(config, state, value1.UnevaluatedItems, value2.UnevaluatedItems)
	if err != nil {
		return nil, err
	}
	result.UnevaluatedPropertiesDiff, err = getSchemaDiff(config, state, value1.UnevaluatedProperties, value2.UnevaluatedProperties)
	if err != nil {
		return nil, err
	}
	result.IfDiff, err = getSchemaDiff(config, state, value1.If, value2.If)
	if err != nil {
		return nil, err
	}
	result.ThenDiff, err = getSchemaDiff(config, state, value1.Then, value2.Then)
	if err != nil {
		return nil, err
	}
	result.ElseDiff, err = getSchemaDiff(config, state, value1.Else, value2.Else)
	if err != nil {
		return nil, err
	}
	result.DependentRequiredDiff = getValueDiff(value1.DependentRequired, value2.DependentRequired)
	result.SchemaIDDiff = getValueDiff(value1.SchemaID, value2.SchemaID)
	result.AnchorDiff = getValueDiff(value1.Anchor, value2.Anchor)
	result.DynamicRefDiff = getValueDiff(value1.DynamicRef, value2.DynamicRef)
	result.DynamicAnchorDiff = getValueDiff(value1.DynamicAnchor, value2.DynamicAnchor)
	result.ContentMediaTypeDiff = getValueDiff(value1.ContentMediaType, value2.ContentMediaType)
	result.ContentEncodingDiff = getValueDiff(value1.ContentEncoding, value2.ContentEncoding)
	result.ContentSchemaDiff, err = getSchemaDiff(config, state, value1.ContentSchema, value2.ContentSchema)
	if err != nil {
		return nil, err
	}
	result.DefsDiff, err = getSchemasDiff(config, state, value1.Defs, value2.Defs)
	if err != nil {
		return nil, err
	}
	result.SchemaDialectDiff = getValueDiff(value1.SchemaDialect, value2.SchemaDialect)
	result.CommentDiff = getValueDiff(value1.Comment, value2.Comment)

	return &result, nil
}
