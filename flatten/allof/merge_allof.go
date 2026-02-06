package allof

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	FormatErrorMessage  = "unable to resolve Format conflict using default resolver: all Format values must be identical"
	TypeErrorMessage    = "unable to resolve Type conflict: all Type values must be identical"
	DefaultErrorMessage = "unable to resolve Default conflict: all Default values must be identical"

	FormatInt32  = "int32"
	FormatInt64  = "int64"
	FormatFloat  = "float"
	FormatDouble = "double"
)

type SchemaCollection struct {
	Not                  []*openapi3.SchemaRef
	OneOf                []openapi3.SchemaRefs
	AnyOf                []openapi3.SchemaRefs
	Title                []string
	Type                 []string
	Format               []string
	Description          []string
	Enum                 [][]any
	UniqueItems          []bool
	ExclusiveMin         []openapi3.ExclusiveBound
	ExclusiveMax         []openapi3.ExclusiveBound
	Min                  []*float64
	Max                  []*float64
	MultipleOf           []*float64
	MinLength            []uint64
	MaxLength            []*uint64
	Pattern              []string
	MinItems             []uint64
	MaxItems             []*uint64
	Items                []*openapi3.SchemaRef
	Required             [][]string
	Properties           []openapi3.Schemas
	MinProps             []uint64
	MaxProps             []*uint64
	AdditionalProperties []openapi3.AdditionalProperties
	Nullable             []bool
	ReadOnly             []bool
	WriteOnly            []bool
	Default              []any
}

type state struct {

	// maps original schemas to their merged result schema.
	mergedSchemas map[*openapi3.Schema]*openapi3.Schema

	// indicates whether a reference is circular (true) or non-circular.
	refs map[string]bool

	// after mergeInternal is executed, circularAllOf contains all SchemaRefs which have circular allof.
	circularAllOf openapi3.SchemaRefs
}

func newState() *state {
	return &state{
		mergedSchemas: map[*openapi3.Schema]*openapi3.Schema{},
		refs:          map[string]bool{},
		circularAllOf: openapi3.SchemaRefs{},
	}
}

func Merge(schema openapi3.SchemaRef) (*openapi3.Schema, error) {
	state := newState()
	result, err := mergeInternal(state, &schema)
	if err != nil {
		return nil, err
	}

	if len(state.circularAllOf) == 0 {
		return result.Value, nil
	}

	for _, schema := range state.circularAllOf {
		err := mergeCircularAllOf(state, schema)
		if err != nil {
			return nil, err
		}
	}

	return result.Value, nil
}

func mergeCircularAllOf(state *state, baseSchemaRef *openapi3.SchemaRef) error {
	allOfCopy := make(openapi3.SchemaRefs, len(baseSchemaRef.Value.AllOf))
	copy(allOfCopy, baseSchemaRef.Value.AllOf)

	schemaRefs := openapi3.SchemaRefs{baseSchemaRef}
	schemaRefs = append(schemaRefs, baseSchemaRef.Value.AllOf...)
	err := flattenSchemas(state, baseSchemaRef, schemaRefs)
	if err != nil {
		return err
	}
	baseSchemaRef.Value.AllOf = nil
	pruneOneOf(state, baseSchemaRef, allOfCopy)
	pruneAnyOf(baseSchemaRef)
	return nil
}

func pruneAnyOf(schema *openapi3.SchemaRef) {
	if len(schema.Value.AnyOf) == 1 && schema.Value.AnyOf[0].Value == schema.Value {
		schema.Value.AnyOf = nil
	}
}

// pruneCircularOneOfInHierarchy prunes the 'oneOf' field from a merged schema when specific conditions are met.
// Pruning criteria:
// - The unmerged schema is a child of another parent schema, through the oneOf field.
// - The unmerged schema contains an 'allOf' field with a circular reference to the parent schema.
// - The merged parent and the merged child schemas contain an identical oneOf field.
// - The merged parent schema contains a non-empty propertyName discriminator field.
func pruneCircularOneOfInHierarchy(state *state, merged *openapi3.SchemaRef, allOf openapi3.SchemaRefs) {
	for _, allOfSchema := range allOf {
		isCircular := state.refs[allOfSchema.Ref]
		if !isCircular {
			continue
		}

		// check if merged is a child of allOfSchemna
		isChild := false
		for _, of := range allOfSchema.Value.OneOf {
			if of.Value == merged.Value {
				isChild = true
			}
		}

		if !isChild {
			continue
		}

		if allOfSchema.Value.Discriminator == nil || allOfSchema.Value.Discriminator.PropertyName == "" {
			continue
		}

		if len(allOfSchema.Value.OneOf) != len(merged.Value.OneOf) {
			continue
		}

		// check if oneOf field of allOfSchema matches the oneOf field of merged
		mismatchFound := false
		for i, of := range allOfSchema.Value.OneOf {
			if of.Value != merged.Value.OneOf[i].Value {
				mismatchFound = true
				break
			}
		}

		if !mismatchFound {
			merged.Value.OneOf = nil
			break
		}
	}
}

func pruneOneOf(state *state, merged *openapi3.SchemaRef, allOf openapi3.SchemaRefs) {
	if len(merged.Value.OneOf) == 1 && merged.Value.OneOf[0].Value == merged.Value {
		merged.Value.OneOf = nil
		return
	}
	pruneCircularOneOfInHierarchy(state, merged, allOf)
}

// Merge replaces objects under AllOf with a flattened equivalent
func mergeInternal(state *state, base *openapi3.SchemaRef) (*openapi3.SchemaRef, error) {
	if base == nil {
		return nil, nil
	}

	// return cached result if this schema has already been merged
	cached, ok := state.mergedSchemas[base.Value]
	if ok {
		return openapi3.NewSchemaRef(base.Ref, cached), nil
	}

	result := openapi3.NewSchemaRef(base.Ref, openapi3.NewSchema())

	// map original schema to result
	state.mergedSchemas[base.Value] = result.Value

	result.Value.Title = base.Value.Title
	result.Value.Type = base.Value.Type
	result.Value.Format = base.Value.Format
	result.Value.Description = base.Value.Description
	result.Value.Type = base.Value.Type
	result.Value.Enum = base.Value.Enum
	result.Value.UniqueItems = base.Value.UniqueItems
	result.Value.ExclusiveMax = base.Value.ExclusiveMax
	result.Value.ExclusiveMin = base.Value.ExclusiveMin
	result.Value.Nullable = base.Value.Nullable
	result.Value.ReadOnly = base.Value.ReadOnly
	result.Value.WriteOnly = base.Value.WriteOnly
	result.Value.Min = base.Value.Min
	result.Value.Max = base.Value.Max
	result.Value.MultipleOf = base.Value.MultipleOf
	result.Value.MinLength = base.Value.MinLength
	result.Value.Default = base.Value.Default
	result.Value.Discriminator = base.Value.Discriminator
	if base.Value.MaxLength != nil {
		result.Value.MaxLength = openapi3.Ptr(*base.Value.MaxLength)
	}
	result.Value.Pattern = base.Value.Pattern
	result.Value.MinItems = base.Value.MinItems
	if base.Value.MaxItems != nil {
		result.Value.MaxItems = openapi3.Ptr(*base.Value.MaxItems)
	}
	result.Value.Required = base.Value.Required
	result.Value.MinProps = base.Value.MinProps
	if base.Value.MaxProps != nil {
		result.Value.MaxProps = openapi3.Ptr(*base.Value.MaxProps)
	}

	// merge all fields of type SchemaRef
	allOf, err := mergeSchemaRefs(state, base.Value.AllOf)
	if err != nil {
		return nil, err
	}
	result.Value.OneOf, err = mergeSchemaRefs(state, base.Value.OneOf)
	if err != nil {
		return nil, err
	}
	result.Value.Items, err = mergeInternal(state, base.Value.Items)
	if err != nil {
		return nil, err
	}
	result.Value.AnyOf, err = mergeSchemaRefs(state, base.Value.AnyOf)
	if err != nil {
		return nil, err
	}
	result.Value.Not, err = mergeInternal(state, base.Value.Not)
	if err != nil {
		return nil, err
	}
	result.Value.Properties, err = mergeProperties(state, base.Value.Properties)
	if err != nil {
		return nil, err
	}
	result.Value.AdditionalProperties, err = mergeAdditionalProperties(state, base.Value.AdditionalProperties)
	if err != nil {
		return nil, err
	}

	if len(base.Value.AllOf) == 0 {
		return result, nil
	}
	updateRefs(state, base.Value.AllOf)
	if isAllOfCircular(state, base.Value.AllOf) {
		state.circularAllOf = append(state.circularAllOf, result)
		result.Value.AllOf = allOf
		return result, nil
	}

	// flatten merged schemas into a single equivalent schema
	toFlatten := openapi3.SchemaRefs{result}
	toFlatten = append(toFlatten, allOf...)
	err = flattenSchemas(state, result, toFlatten)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func updateRefs(state *state, srefs openapi3.SchemaRefs) {
	for _, s := range srefs {
		if s.Ref != "" {
			_, ok := state.refs[s.Ref]
			if ok {
				continue
			}
			_, ok = state.mergedSchemas[s.Value]
			state.refs[s.Ref] = ok
		}
	}
}

func isAllOfCircular(state *state, srefs openapi3.SchemaRefs) bool {
	for _, s := range srefs {
		if s.Ref == "" {
			continue
		}
		isCircular, ok := state.refs[s.Ref]
		if ok && isCircular {
			return true
		}
	}
	return false
}

func mergeAdditionalProperties(state *state, ap openapi3.AdditionalProperties) (openapi3.AdditionalProperties, error) {
	result := openapi3.AdditionalProperties{}
	if ap.Schema != nil && ap.Schema.Value != nil {
		merged, err := mergeInternal(state, ap.Schema)
		if err != nil {
			return result, err
		}
		result.Schema = merged
	}
	value := true
	ptr := &value
	if ap.Has != nil {
		*ptr = *ap.Has
	} else {
		ptr = nil
	}
	result.Has = ptr
	return result, nil
}

func mergeProperties(state *state, props openapi3.Schemas) (openapi3.Schemas, error) {
	if props == nil {
		return nil, nil
	}
	result := openapi3.Schemas{}
	for k, v := range props {
		merged, err := mergeInternal(state, v)
		if err != nil {
			return nil, err
		}
		result[k] = merged
	}
	return result, nil
}

func mergeSchemaRefs(state *state, srefs openapi3.SchemaRefs) (openapi3.SchemaRefs, error) {
	if srefs == nil {
		return nil, nil
	}
	result := openapi3.SchemaRefs{}
	for _, s := range srefs {
		merged, err := mergeInternal(state, s)
		if err != nil {
			return nil, err
		}
		result = append(result, merged)
	}
	return result, nil
}

// Given a list of schemas that are free of AllOf or nested AllOf components as input,
// the function produces a single equivalent schema in the resultRef parameter.
func flattenSchemas(state *state, result *openapi3.SchemaRef, schemas []*openapi3.SchemaRef) error {

	collection := collect(schemas)
	var err error

	result.Value.Title = firstOrSecondNonEmpty(collection.Title)
	result.Value.Description = firstOrSecondNonEmpty(collection.Description)
	result.Value = resolveNumberRange(result.Value, &collection)
	result.Value.MinLength = findMaxValue(collection.MinLength)
	result.Value.MaxLength = findMinValue(collection.MaxLength)
	result.Value.MinItems = findMaxValue(collection.MinItems)
	result.Value.MaxItems = findMinValue(collection.MaxItems)
	result.Value.MinProps = findMaxValue(collection.MinProps)
	result.Value.MaxProps = findMinValue(collection.MaxProps)
	result.Value.Pattern = resolvePattern(collection.Pattern)
	result.Value.Nullable = !hasFalse(collection.Nullable)
	result.Value.ReadOnly = hasTrue(collection.ReadOnly)
	result.Value.WriteOnly = hasTrue(collection.WriteOnly)
	result.Value.Required = resolveRequired(collection.Required)
	result.Value = resolveMultipleOf(result.Value, &collection)
	result.Value.UniqueItems = resolveUniqueItems(collection.UniqueItems)
	result.Value.Default, err = resolveDefault(&collection)
	if err != nil {
		return err
	}
	result.Value.Enum, err = resolveEnum(collection.Enum)
	if err != nil {
		return err
	}
	result.Value, err = resolveFormat(result.Value, &collection)
	if err != nil {
		return err
	}
	result.Value, err = resolveType(result.Value, &collection)
	if err != nil {
		return err
	}
	result.Value, err = resolveItems(state, result.Value, &collection)
	if err != nil {
		return err
	}
	result.Value, err = resolveProperties(state, result.Value, &collection)
	if err != nil {
		return err
	}
	result.Value, err = resolveOneOf(state, result.Value, &collection)
	if err != nil {
		return err
	}
	result.Value, err = resolveAnyOf(state, result.Value, &collection)
	if err != nil {
		return err
	}
	result.Value, err = resolveNot(state, result.Value, &collection)
	return err
}

func hasTrue(values []bool) bool {
	for _, val := range values {
		if val {
			return true
		}
	}
	return false
}

func hasFalse(values []bool) bool {
	for _, val := range values {
		if !val {
			return true
		}
	}
	return false
}

func resolveNumberRange(schema *openapi3.Schema, collection *SchemaCollection) *openapi3.Schema {

	//resolve minimum
	max := math.Inf(-1)
	var exclusiveMin openapi3.ExclusiveBound
	var value *float64
	for i, s := range collection.Min {
		if s != nil {
			if *s > max {
				max = *s
				value = s
				exclusiveMin = collection.ExclusiveMin[i]
			}
		}
	}

	schema.Min = value
	schema.ExclusiveMin = exclusiveMin
	//resolve maximum
	min := math.Inf(1)
	var exclusiveMax openapi3.ExclusiveBound
	// var value *float64
	for i, s := range collection.Max {
		if s != nil {
			if *s < min {
				min = *s
				value = s
				exclusiveMax = collection.ExclusiveMax[i]
			}
		}
	}

	schema.Max = value
	schema.ExclusiveMax = exclusiveMax
	return schema
}

func resolveItems(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {

	items := openapi3.SchemaRefs{}
	for _, sref := range collection.Items {
		if sref != nil {
			items = append(items, sref)
		}
	}
	if len(items) == 0 {
		schema.Items = nil
		return schema, nil
	}
	if len(items) == 1 {
		schema.Items = items[0]
		return schema, nil
	}
	result := openapi3.NewSchemaRef("", openapi3.NewSchema())
	err := flattenSchemas(state, result, items)
	if err != nil {
		return nil, err
	}
	schema.Items = result
	return schema, nil
}

func resolveUniqueItems(values []bool) bool {
	for _, v := range values {
		if v {
			return true
		}
	}
	return false
}

/* MultipleOf */
func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b uint64) uint64 {
	return a * b / gcd(a, b)
}

func containsNonInteger(arr []float64) bool {
	for _, num := range arr {
		if num != math.Trunc(num) {
			return true
		}
	}
	return false
}

func resolveMultipleOf(schema *openapi3.Schema, collection *SchemaCollection) *openapi3.Schema {
	values := []float64{}
	for _, v := range collection.MultipleOf {
		if v == nil {
			continue
		}
		values = append(values, *v)
	}
	if len(values) == 0 {
		schema.MultipleOf = nil
		return schema
	}

	factor := 1.0
	for containsNonInteger(values) {
		factor *= 10.0
		for i := range values {
			values[i] *= factor
		}
	}
	uintValues := make([]uint64, len(values))
	for i, val := range values {
		uintValues[i] = uint64(val)
	}
	lcmValue := uintValues[0]
	for _, v := range uintValues {
		lcmValue = lcm(lcmValue, v)
	}
	schema.MultipleOf = openapi3.Ptr(float64(lcmValue) / factor)
	return schema
}

func hasFalseValue(ap []openapi3.AdditionalProperties) bool {
	for _, v := range ap {
		if v.Has != nil && !*v.Has {
			return true
		}
	}
	return false
}

// resolve properties which have additionalProperties that are set to false.
func resolveFalseProps(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	schema = resolveFalseAdditionalProps(schema, collection)
	propsToMerge := getFalsePropsKeys(collection)
	return mergeProps(state, schema, collection, propsToMerge)
}

func resolveFalseAdditionalProps(schema *openapi3.Schema, collection *SchemaCollection) *openapi3.Schema {
	has := false
	schema.AdditionalProperties.Has = &has
	return schema
}

// if there are additionalProperties which are Schemas, they are merged to a single Schema.

func resolveNonFalseAdditionalProps(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	additionalSchemas := openapi3.SchemaRefs{}
	for _, ap := range collection.AdditionalProperties {
		if ap.Schema != nil {
			additionalSchemas = append(additionalSchemas, ap.Schema)
		}
	}

	var schemaRef *openapi3.SchemaRef
	if len(additionalSchemas) == 1 {
		schemaRef = additionalSchemas[0]
	} else if len(additionalSchemas) > 1 {
		result := openapi3.NewSchemaRef("", openapi3.NewSchema())
		err := flattenSchemas(state, result, additionalSchemas)
		if err != nil {
			return nil, err
		}
		schemaRef = result
	}
	schema.AdditionalProperties.Has = nil
	schema.AdditionalProperties.Schema = schemaRef
	return schema, nil
}

func resolveNonFalseProps(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	result, err := resolveNonFalseAdditionalProps(state, schema, collection)
	if err != nil {
		return nil, err
	}
	propsToMerge := getNonFalsePropsKeys(collection)
	return mergeProps(state, result, collection, propsToMerge)
}

// the output is the intersection of all properties keys of schemas which have additionalProperties set to false.
func getFalsePropsKeys(collection *SchemaCollection) []string {
	properties := [][]string{}
	for i, schema := range collection.Properties {
		additionalProps := collection.AdditionalProperties[i].Has
		if additionalProps != nil && !*additionalProps {
			keys := []string{}
			for key := range schema {
				keys = append(keys, key)
			}
			properties = append(properties, keys)
		}
	}
	return findIntersection(properties...)
}

// the output is a list of unique properties keys of all schemas.
func getNonFalsePropsKeys(collection *SchemaCollection) []string {
	keys := []string{}
	for _, schema := range collection.Properties {
		for key := range schema {
			keys = append(keys, key)
		}
	}
	return getUniqueStrings(keys)
}

func getUniqueStrings(input []string) []string {
	uniqueStrings := make(map[string]bool)

	for _, str := range input {
		uniqueStrings[str] = true
	}

	result := []string{}
	for str := range uniqueStrings {
		result = append(result, str)
	}

	return result
}

func resolveProperties(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	if hasFalseValue(collection.AdditionalProperties) {
		return resolveFalseProps(state, schema, collection)
	} else {
		return resolveNonFalseProps(state, schema, collection)
	}
}

func mergeProps(state *state, schema *openapi3.Schema, collection *SchemaCollection, propsToMerge []string) (*openapi3.Schema, error) {
	propsToSchemasMap := map[string]openapi3.SchemaRefs{}
	for _, schema := range collection.Properties {
		for propKey, schemaRef := range schema {
			if containsString(propsToMerge, propKey) {
				propsToSchemasMap[propKey] = append(propsToSchemasMap[propKey], schemaRef)
			}
		}
	}

	result := make(openapi3.Schemas)
	for prop, schemas := range propsToSchemasMap {
		flattened := openapi3.NewSchemaRef("", openapi3.NewSchema())
		err := flattenSchemas(state, flattened, schemas)
		if err != nil {
			return nil, err
		}
		result[prop] = flattened
	}

	if len(result) == 0 {
		result = nil
	}

	schema.Properties = result
	return schema, nil
}

func resolveEnum(values [][]any) ([]any, error) {
	var nonEmptyEnum [][]any
	for _, enum := range values {
		if len(enum) > 0 {
			nonEmptyEnum = append(nonEmptyEnum, enum)
		}
	}
	var intersection []any
	if len(nonEmptyEnum) == 0 {
		return intersection, nil
	}
	intersection = findIntersectionOfArrays(nonEmptyEnum)
	if len(intersection) == 0 {
		return nil, errors.New("unable to resolve Enum conflict: intersection of values must be non-empty")
	}
	return intersection, nil
}

func resolvePattern(values []string) string {
	patterns := []string{}
	for _, v := range values {
		if len(v) > 0 {
			patterns = append(patterns, v)
		}
	}
	if len(patterns) == 0 {
		return ""
	}
	if len(patterns) == 1 {
		return patterns[0]
	}
	var pattern strings.Builder
	for _, p := range patterns {
		if len(p) > 0 {
			if !isPatternResolved(p) {
				pattern.WriteString(fmt.Sprintf("(?=%s)", p))
			} else {
				pattern.WriteString(p)
			}
		}
	}
	return pattern.String()
}

func isPatternResolved(pattern string) bool {
	match, _ := regexp.MatchString(`^\(\?=.+\)$`, pattern)
	return match
}

func findMaxValue(values []uint64) uint64 {
	max := uint64(0)
	for _, num := range values {
		if num > max {
			max = num
		}
	}
	return max
}

func findMinValue(values []*uint64) *uint64 {
	dvalues := []uint64{}
	for _, v := range values {
		if v != nil {
			dvalues = append(dvalues, *v)
		}
	}
	if len(dvalues) == 0 {
		return nil
	}
	min := uint64(math.MaxUint64)
	for _, num := range dvalues {
		if num < min {
			min = num
		}
	}
	return openapi3.Ptr(min)
}

func resolveType(schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	types := filterEmptyStrings(collection.Type)
	if len(types) == 0 {
		schema.Type = nil
		return schema, nil
	}
	if areTypesNumeric(types) {
		for _, t := range types {
			if t == "integer" {
				schema.Type = &openapi3.Types{"integer"}
				return schema, nil
			}
		}
		schema.Type = &openapi3.Types{"number"}
		return schema, nil
	}
	if allStringsEqual(types) {
		schema.Type = &openapi3.Types{types[0]}
		return schema, nil
	}
	return schema, errors.New(TypeErrorMessage)
}

func areTypesNumeric(types []string) bool {
	for _, t := range types {
		if t != "integer" && t != "number" {
			return false
		}
	}
	return true
}

func resolveFormat(schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	formats := filterEmptyStrings(collection.Format)
	if len(formats) == 0 {
		schema.Format = ""
		return schema, nil
	}
	if areFormatsNumeric(formats) {
		schema.Format = resolveNumericFormat(formats)
		return schema, nil
	}
	return defaultFormatResolver(schema, formats)
}

func resolveNumericFormat(formats []string) string {
	orderMap := make(map[string]int)
	orderMap[FormatInt32] = 1
	orderMap[FormatInt64] = 2
	orderMap[FormatFloat] = 3
	orderMap[FormatDouble] = 4
	result := FormatDouble
	for _, format := range formats {
		if orderMap[format] < orderMap[result] {
			result = format
		}
	}
	return result
}

func defaultFormatResolver(schema *openapi3.Schema, formats []string) (*openapi3.Schema, error) {
	if allStringsEqual(formats) {
		schema.Format = formats[0]
		return schema, nil
	}
	return &openapi3.Schema{}, errors.New(FormatErrorMessage)
}

func areFormatsNumeric(values []string) bool {
	for _, val := range values {
		if val != FormatInt32 && val != FormatInt64 && val != FormatFloat && val != FormatDouble {
			return false
		}
	}
	return true
}

func containsString(list []string, search string) bool {
	for _, item := range list {
		if item == search {
			return true
		}
	}
	return false
}

func filterEmptyStrings(input []string) []string {
	var result []string

	for _, s := range input {
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}

func allStringsEqual(values []string) bool {
	first := values[0]
	for _, value := range values {
		if first != value {
			return false
		}
	}
	return true
}

func getIntersection(arr1, arr2 []any) []any {
	intersectionMap := make(map[any]bool)
	result := []any{}

	// Mark elements in the first array
	for _, val := range arr1 {
		intersectionMap[val] = true
	}

	// Check if elements in the second array exist in the intersection map
	for _, val := range arr2 {
		if intersectionMap[val] {
			result = append(result, val)
		}
	}

	return result
}

func findIntersectionOfArrays(arrays [][]any) []any {
	if len(arrays) == 0 {
		return nil
	}

	intersection := arrays[0]

	for i := 1; i < len(arrays); i++ {
		intersection = getIntersection(intersection, arrays[i])
	}
	if len(intersection) == 0 {
		return nil
	}
	return intersection
}

func flattenArray(arrays [][]string) []string {
	var result []string

	for i := 0; i < len(arrays); i++ {
		for j := 0; j < len(arrays[i]); j++ {
			result = append(result, arrays[i][j])
		}
	}

	return result
}

func resolveRequired(values [][]string) []string {
	flatValues := flattenArray(values)
	uniqueMap := make(map[string]bool)
	var uniqueValues []string
	for _, str := range flatValues {
		if _, found := uniqueMap[str]; !found {
			uniqueMap[str] = true
			uniqueValues = append(uniqueValues, str)
		}
	}
	return uniqueValues
}

func collect(schemas []*openapi3.SchemaRef) SchemaCollection {
	collection := SchemaCollection{}
	for _, s := range schemas {
		if s == nil {
			continue
		}
		collection.Not = append(collection.Not, s.Value.Not)
		collection.AnyOf = append(collection.AnyOf, s.Value.AnyOf)
		collection.OneOf = append(collection.OneOf, s.Value.OneOf)
		collection.Title = append(collection.Title, s.Value.Title)
		if s.Value.Type != nil {
			collection.Type = append(collection.Type, *s.Value.Type...)
		}
		collection.Format = append(collection.Format, s.Value.Format)
		collection.Description = append(collection.Description, s.Value.Description)
		collection.Enum = append(collection.Enum, s.Value.Enum)
		collection.UniqueItems = append(collection.UniqueItems, s.Value.UniqueItems)
		collection.ExclusiveMin = append(collection.ExclusiveMin, s.Value.ExclusiveMin)
		collection.ExclusiveMax = append(collection.ExclusiveMax, s.Value.ExclusiveMax)
		collection.Min = append(collection.Min, s.Value.Min)
		collection.Max = append(collection.Max, s.Value.Max)
		collection.MultipleOf = append(collection.MultipleOf, s.Value.MultipleOf)
		collection.MinLength = append(collection.MinLength, s.Value.MinLength)
		collection.MaxLength = append(collection.MaxLength, s.Value.MaxLength)
		collection.Pattern = append(collection.Pattern, s.Value.Pattern)
		collection.MinItems = append(collection.MinItems, s.Value.MinItems)
		collection.MaxItems = append(collection.MaxItems, s.Value.MaxItems)
		collection.Items = append(collection.Items, s.Value.Items)
		collection.Required = append(collection.Required, s.Value.Required)
		collection.Properties = append(collection.Properties, s.Value.Properties)
		collection.MinProps = append(collection.MinProps, s.Value.MinProps)
		collection.MaxProps = append(collection.MaxProps, s.Value.MaxProps)
		collection.AdditionalProperties = append(collection.AdditionalProperties, s.Value.AdditionalProperties)
		collection.Nullable = append(collection.Nullable, s.Value.Nullable)
		collection.ReadOnly = append(collection.ReadOnly, s.Value.ReadOnly)
		collection.WriteOnly = append(collection.WriteOnly, s.Value.WriteOnly)
		collection.Default = append(collection.Default, s.Value.Default)
	}
	return collection
}

// getCombinations calculates the cartesian product of groups of SchemaRefs.
func getCombinations(groups []openapi3.SchemaRefs) []openapi3.SchemaRefs {
	if len(groups) == 0 {
		return []openapi3.SchemaRefs{}
	}
	result := []openapi3.SchemaRefs{{}}
	for _, group := range groups {
		var newResult []openapi3.SchemaRefs
		for _, resultItem := range result {
			for _, ref := range group {
				combination := append(openapi3.SchemaRefs{}, resultItem...)
				combination = append(combination, ref)
				newResult = append(newResult, combination)
			}
		}
		result = newResult
	}
	return result
}

func flattenCombinations(state *state, combinations []openapi3.SchemaRefs) ([]*openapi3.SchemaRef, error) {
	flattened := []*openapi3.SchemaRef{}
	for _, combination := range combinations {
		result := openapi3.NewSchemaRef("", openapi3.NewSchema())
		err := flattenSchemas(state, result, combination)
		if err != nil {
			continue
		}
		flattened = append(flattened, result)
	}
	if len(flattened) == 0 {
		return nil, errors.New("unable to resolve combined schema")
	}
	return flattened, nil
}

func resolveNot(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	result := filterNilSchemaRef(collection.Not)
	if len(result) == 0 {
		return schema, nil
	}
	if len(result) == 1 {
		schema.Not = result[0]
		return schema, nil
	}
	schema.Not = openapi3.NewSchemaRef("", &openapi3.Schema{
		AnyOf: result,
	})
	return schema, nil
}

func filterNilSchemaRef(refs openapi3.SchemaRefs) openapi3.SchemaRefs {
	result := []*openapi3.SchemaRef{}
	for _, v := range refs {
		if v != nil {
			result = append(result, v)
		}
	}
	return result
}

func filterEmptySchemaRefs(groups []openapi3.SchemaRefs) []openapi3.SchemaRefs {
	result := []openapi3.SchemaRefs{}
	for _, group := range groups {
		if len(group) > 0 {
			result = append(result, group)
		}
	}
	return result
}

// deduplicateSchemaRefsGroups removes duplicate groups from the collection.
// Two groups are considered equivalent if they contain the same set of schema references.
// This prevents unnecessary cartesian product expansion when multiple allOf subschemas
// reference the same oneOf/anyOf definition.
// Note: caller (resolveCombinations) already handles len(groups) <= 1 cases.
func deduplicateSchemaRefsGroups(groups []openapi3.SchemaRefs) []openapi3.SchemaRefs {
	seen := make(map[string]bool)
	result := []openapi3.SchemaRefs{}

	for _, group := range groups {
		key := schemaRefsGroupKey(group)
		if !seen[key] {
			seen[key] = true
			result = append(result, group)
		}
	}

	return result
}

// schemaRefsGroupKey generates a unique key for a group of SchemaRefs.
// For referenced schemas, uses the $ref path.
// For inline schemas, uses a string representation of the schema content.
// Note: caller ensures refs is non-empty (via filterEmptySchemaRefs).
func schemaRefsGroupKey(refs openapi3.SchemaRefs) string {
	keys := make([]string, len(refs))
	for i, ref := range refs {
		if ref.Ref != "" {
			keys[i] = ref.Ref
		} else {
			keys[i] = fmt.Sprintf("%v", ref.Value)
		}
	}

	// Sort to ensure order-independent comparison
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	for i := 0; i < len(sortedKeys)-1; i++ {
		for j := i + 1; j < len(sortedKeys); j++ {
			if sortedKeys[i] > sortedKeys[j] {
				sortedKeys[i], sortedKeys[j] = sortedKeys[j], sortedKeys[i]
			}
		}
	}

	return strings.Join(sortedKeys, "|")
}

func resolveAnyOf(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	refs, err := resolveCombinations(state, collection.AnyOf)
	schema.AnyOf = refs
	return schema, err
}

func resolveOneOf(state *state, schema *openapi3.Schema, collection *SchemaCollection) (*openapi3.Schema, error) {
	refs, err := resolveCombinations(state, collection.OneOf)
	schema.OneOf = refs
	return schema, err
}

func resolveCombinations(state *state, collection []openapi3.SchemaRefs) (openapi3.SchemaRefs, error) {
	groups := filterEmptySchemaRefs(collection)
	if len(groups) == 0 {
		return nil, nil
	}

	// there is only one group of schemas, no need for calculating combinations.
	if len(groups) == 1 {
		return groups[0], nil
	}

	// deduplicate equivalent groups to avoid unnecessary cartesian product expansion
	// e.g., two identical oneOf: [A, B] should not produce [A,A], [A,B], [B,A], [B,B]
	groups = deduplicateSchemaRefsGroups(groups)
	if len(groups) == 1 {
		return groups[0], nil
	}

	combinations := getCombinations(groups)
	flattenedCombinations, err := flattenCombinations(state, combinations)
	if err != nil {
		return nil, err
	}
	return flattenedCombinations, nil
}

func findIntersection(arrays ...[]string) []string {
	if len(arrays) == 0 {
		return nil
	}

	// Create a map to store the elements of the first array
	elementsMap := make(map[string]bool)
	for _, element := range arrays[0] {
		elementsMap[element] = true
	}

	// Iterate through the remaining arrays and update the map
	for _, arr := range arrays[1:] {
		tempMap := make(map[string]bool)
		for _, element := range arr {
			if elementsMap[element] {
				tempMap[element] = true
			}
		}
		elementsMap = tempMap
	}

	intersection := []string{}
	for element := range elementsMap {
		intersection = append(intersection, element)
	}

	return intersection
}

func resolveDefault(collection *SchemaCollection) (any, error) {
	values := make([]any, 0)
	for _, v := range collection.Default {
		if v != nil {
			values = append(values, v)
		}
	}
	if len(values) == 0 {
		return nil, nil
	}
	first := values[0]
	for _, v := range values {
		if !reflect.DeepEqual(first, v) {
			return nil, errors.New(DefaultErrorMessage)
		}
	}
	return first, nil
}

// FirstOrSecondNonEmpty returns the first non-empty string from the first two elements of a slice.
// If both are empty or the slice is empty, it returns an empty string.
func firstOrSecondNonEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	if len(values) == 1 {
		return values[0]
	}
	if values[0] != "" {
		return values[0]
	}
	return values[1]
}
