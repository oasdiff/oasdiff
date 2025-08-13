package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

// ParametersDiffByLocation describes the changes, grouped by param location, between a pair of lists of parameter objects: https://swagger.io/specification/#parameter-object
type ParametersDiffByLocation struct {
	Added    ParamNamesByLocation `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  ParamNamesByLocation `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ParamDiffByLocation  `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *ParametersDiffByLocation) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

// ParamLocations are the four possible locations of parameters: path, query, header or cookie
var ParamLocations = []string{openapi3.ParameterInPath, openapi3.ParameterInQuery, openapi3.ParameterInHeader, openapi3.ParameterInCookie}

// ParamNamesByLocation maps param location (path, query, header or cookie) to the params in this location
type ParamNamesByLocation map[string]utils.StringList

// Len returns the number of all params in all locations
func (params ParamNamesByLocation) Len() int {
	return lenNested(params)
}

// ParamDiffByLocation maps param location (path, query, header or cookie) to param diffs in this location
type ParamDiffByLocation map[string]ParamDiffs

// Len returns the number of all params in all locations
func (params ParamDiffByLocation) Len() int {
	return lenNested(params)
}

func lenNested[T utils.StringList | ParamDiffs](mapOfList map[string]T) int {
	result := 0
	for _, l := range mapOfList {
		result += len(l)
	}
	return result
}

func newParametersDiffByLocation() *ParametersDiffByLocation {
	return &ParametersDiffByLocation{
		Added:    ParamNamesByLocation{},
		Deleted:  ParamNamesByLocation{},
		Modified: ParamDiffByLocation{},
	}
}

// ParamDiffs is map of parameter names to their respective diffs
type ParamDiffs map[string]*ParameterDiff

func (diff *ParametersDiffByLocation) addAddedParam(param *openapi3.Parameter) {

	if paramNames, ok := diff.Added[param.In]; ok {
		diff.Added[param.In] = append(paramNames, param.Name)
	} else {
		diff.Added[param.In] = utils.StringList{param.Name}
	}
}

func (diff *ParametersDiffByLocation) addDeletedParam(param *openapi3.Parameter) {

	if paramNames, ok := diff.Deleted[param.In]; ok {
		diff.Deleted[param.In] = append(paramNames, param.Name)
	} else {
		diff.Deleted[param.In] = utils.StringList{param.Name}
	}
}

func (diff *ParametersDiffByLocation) addModifiedParam(param *openapi3.Parameter, paramDiff *ParameterDiff) {

	if paramDiffs, ok := diff.Modified[param.In]; ok {
		paramDiffs[param.Name] = paramDiff
	} else {
		diff.Modified[param.In] = ParamDiffs{param.Name: paramDiff}
	}
}

func getParametersDiffByLocation(config *Config, state *state, params1, params2 openapi3.Parameters, pathParamsMap PathParamsMap) (*ParametersDiffByLocation, error) {
	diff, err := getParametersDiffByLocationInternal(config, state, params1, params2, pathParamsMap)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getParametersDiffByLocationInternal(config *Config, state *state, params1, params2 openapi3.Parameters, pathParamsMap PathParamsMap) (*ParametersDiffByLocation, error) {

	result := newParametersDiffByLocation()

	// Track which parameters have been handled by exploded parameter matching
	processedParams1 := make(map[*openapi3.Parameter]bool)
	processedParams2 := make(map[*openapi3.Parameter]bool)

	// First pass: Handle exploded parameter equivalences
	err := handleExplodedParameterMatching(config, state, params1, params2, result, processedParams1, processedParams2)
	if err != nil {
		return nil, err
	}

	// Second pass: Handle regular parameter matching for remaining parameters
	for _, paramRef1 := range params1 {
		param1, err := derefParam(paramRef1)
		if err != nil {
			return nil, err
		}

		// Skip if already processed by exploded parameter matching
		if processedParams1[param1] {
			continue
		}

		param2, err := findParam(param1, params2, pathParamsMap)
		if err != nil {
			return nil, err
		}

		if param2 != nil && !processedParams2[param2] {
			diff, err := getParameterDiff(config, state, param1, param2)
			if err != nil {
				return nil, err
			}

			if !diff.Empty() {
				result.addModifiedParam(param1, diff)
			}
			processedParams2[param2] = true
		} else {
			result.addDeletedParam(param1)
		}
	}

	// Third pass: Handle added parameters
	for _, paramRef2 := range params2 {
		param2, err := derefParam(paramRef2)
		if err != nil {
			return nil, err
		}

		// Skip if already processed
		if processedParams2[param2] {
			continue
		}

		result.addAddedParam(param2)
	}

	return result, nil
}

func derefParam(ref *openapi3.ParameterRef) (*openapi3.Parameter, error) {

	if ref == nil || ref.Value == nil {
		return nil, fmt.Errorf("parameter reference is nil")
	}

	return ref.Value, nil
}

// findParam looks for a param that matches param1 in params2 taking into account param renaming through pathParamsMap
func findParam(param1 *openapi3.Parameter, params2 openapi3.Parameters, pathParamsMap PathParamsMap) (*openapi3.Parameter, error) {
	// TODO: optimize with a map
	for _, paramRef2 := range params2 {
		param2, err := derefParam(paramRef2)
		if err != nil {
			return nil, err
		}

		equal, err := equalParams(param1, param2, pathParamsMap)
		if err != nil {
			return nil, err
		}

		if equal {
			return param2, nil
		}
	}

	return nil, nil
}

func equalParams(param1 *openapi3.Parameter, param2 *openapi3.Parameter, pathParamsMap PathParamsMap) (bool, error) {
	if param1 == nil || param2 == nil {
		return false, fmt.Errorf("param is nil")
	}

	if param1.In != param2.In {
		return false, nil
	}

	if param1.In != openapi3.ParameterInPath {
		// Simple name comparison for non-path parameters
		if param1.Name == param2.Name {
			return true, nil
		}

		// Check for semantic equivalence with exploded parameters
		return isSemanticEquivalent(param1, param2), nil
	}

	return pathParamsMap.find(param1.Name, param2.Name), nil
}

func (diff *ParametersDiffByLocation) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}

// isSemanticEquivalent checks if two parameters are semantically equivalent,
// considering the case where separate parameters are consolidated into an exploded object parameter
func isSemanticEquivalent(param1, param2 *openapi3.Parameter) bool {
	// Check if param1 is a simple parameter and param2 is an exploded object
	if isExplodedObjectParam(param2) {
		return isParamInExplodedObject(param1, param2)
	}

	// Check if param2 is a simple parameter and param1 is an exploded object
	if isExplodedObjectParam(param1) {
		return isParamInExplodedObject(param2, param1)
	}

	return false
}

// isExplodedObjectParam checks if a parameter is an exploded object parameter
// (style=form, explode=true, object schema)
func isExplodedObjectParam(param *openapi3.Parameter) bool {
	if param == nil || param.Schema == nil || param.Schema.Value == nil {
		return false
	}

	// Must be explode=true and style=form (or default for query/cookie params)
	explode := param.Explode != nil && *param.Explode

	// Check if style is form, or if it's empty and this is a query/cookie parameter
	// where form is the default style according to OpenAPI spec
	var isFormStyle bool
	switch param.Style {
	case openapi3.SerializationForm:
		isFormStyle = true
	case "":
		// Empty style defaults to "form" only for query and cookie parameters
		// Path and header parameters default to "simple"
		isFormStyle = isQueryOrCookieParam(param)
	}

	// Must have object schema with properties
	schema := param.Schema.Value
	isObjectWithProps := schema.Type != nil && len(*schema.Type) == 1 && (*schema.Type)[0] == "object" && len(schema.Properties) > 0

	return explode && isFormStyle && isObjectWithProps
}

// isParamInExplodedObject checks if a simple parameter corresponds to a property
// in an exploded object parameter
func isParamInExplodedObject(simpleParam, explodedParam *openapi3.Parameter) bool {
	if !isExplodedObjectParam(explodedParam) {
		return false
	}

	// Parameters must be in the same location (query, header, etc.)
	if simpleParam.In != explodedParam.In {
		return false
	}

	schema := explodedParam.Schema.Value
	if schema == nil || schema.Properties == nil {
		return false
	}

	// Check if the simple parameter's name matches a property in the object
	_, exists := schema.Properties[simpleParam.Name]
	return exists
}

// getPropertyDiff compares a simple parameter with its corresponding property in an exploded object parameter.
// The direction parameter controls the diff direction:
// - true: simple → exploded (shows changes from simple to exploded parameter style)
// - false: exploded → simple (shows changes from exploded to simple parameter style)
func getPropertyDiff(config *Config, state *state, simpleParam *openapi3.Parameter, explodedParam *openapi3.Parameter, simpleToExploded bool) (*ParameterDiff, error) {
	if !isExplodedObjectParam(explodedParam) {
		return nil, fmt.Errorf("exploded parameter is not a valid exploded object parameter")
	}

	// Get the property schema from the exploded object
	propertyName := simpleParam.Name
	explodedSchema := explodedParam.Schema.Value
	propertySchemaRef, exists := explodedSchema.Properties[propertyName]
	if !exists {
		return nil, fmt.Errorf("property %s not found in exploded parameter", propertyName)
	}

	// Create a virtual parameter representing what the exploded property would look like
	// as a standalone parameter
	virtualExplodedParam := &openapi3.Parameter{
		Name:     simpleParam.Name,       // Must match for comparison
		In:       simpleParam.In,         // Must match for comparison
		Schema:   propertySchemaRef,      // The key thing we're comparing
		Required: explodedParam.Required, // Use exploded param's required status
		Style:    explodedParam.Style,    // Use exploded param's style
		Explode:  explodedParam.Explode,  // Use exploded param's explode

		// Copy other fields from exploded param
		Description:     explodedParam.Description,
		Deprecated:      explodedParam.Deprecated,
		AllowEmptyValue: explodedParam.AllowEmptyValue,
		AllowReserved:   explodedParam.AllowReserved,
		Example:         explodedParam.Example,
		Examples:        explodedParam.Examples,
		Content:         explodedParam.Content,
		Extensions:      explodedParam.Extensions,
	}

	// Generate diff in the appropriate direction
	if simpleToExploded {
		// Simple → Exploded: compare simple parameter with virtual exploded parameter
		return getParameterDiff(config, state, simpleParam, virtualExplodedParam)
	} else {
		// Exploded → Simple: compare virtual exploded parameter with simple parameter
		return getParameterDiff(config, state, virtualExplodedParam, simpleParam)
	}
}

// matchExplodedWithSimple finds exploded parameters in explodedParams and matches them with
// corresponding simple parameters in simpleParams, processing the diffs accordingly.
func matchExplodedWithSimple(config *Config, state *state, simpleParams, explodedParams openapi3.Parameters, result *ParametersDiffByLocation, processedSimple, processedExploded map[*openapi3.Parameter]bool, simpleToExploded bool) error {

	// Look for exploded parameters in explodedParams
	for _, explodedParamRef := range explodedParams {
		explodedParam, err := derefParam(explodedParamRef)
		if err != nil {
			return err
		}

		if !isExplodedObjectParam(explodedParam) || processedExploded[explodedParam] {
			continue
		}

		// Only apply exploded parameter matching to query and cookie parameters
		// Path and header parameters don't support form style with explode
		if !isQueryOrCookieParam(explodedParam) {
			continue
		}

		// Find all simple parameters that match properties of this exploded parameter
		var matchingParams []*openapi3.Parameter
		for _, simpleParamRef := range simpleParams {
			simpleParam, err := derefParam(simpleParamRef)
			if err != nil {
				return err
			}

			if processedSimple[simpleParam] {
				continue
			}

			// Only match simple parameters that are in query or cookie locations
			// to ensure both simple and exploded params are in compatible locations
			if !isQueryOrCookieParam(simpleParam) {
				continue
			}

			if isParamInExplodedObject(simpleParam, explodedParam) {
				matchingParams = append(matchingParams, simpleParam)
			}
		}

		// If we found matching simple parameters, process them
		if len(matchingParams) > 0 {
			for _, simpleParam := range matchingParams {
				// Generate property-level diff based on direction
				propertyDiff, err := getPropertyDiff(config, state, simpleParam, explodedParam, simpleToExploded)
				if err != nil {
					return err
				}

				if !propertyDiff.Empty() {
					// Use the simple parameter name for reporting
					result.addModifiedParam(simpleParam, propertyDiff)
				}

				// Mark simple parameter as processed
				processedSimple[simpleParam] = true
			}
			// Mark exploded parameter as processed
			processedExploded[explodedParam] = true
		}
	}

	return nil
}

func isQueryOrCookieParam(param *openapi3.Parameter) bool {
	return param.In == openapi3.ParameterInQuery || param.In == openapi3.ParameterInCookie
}

// handleExplodedParameterMatching handles parameter matching between two parameter sets.
// The function automatically detects which parameters are exploded and processes accordingly.
func handleExplodedParameterMatching(config *Config, state *state, params1, params2 openapi3.Parameters, result *ParametersDiffByLocation, processedParams1, processedParams2 map[*openapi3.Parameter]bool) error {

	// Case 1: Simple parameters (params1) → Exploded parameters (params2)
	err := matchExplodedWithSimple(config, state, params1, params2, result, processedParams1, processedParams2, true)
	if err != nil {
		return err
	}

	// Case 2: Exploded parameters (params1) → Simple parameters (params2)
	// Reverse the parameter order to swap the direction
	err = matchExplodedWithSimple(config, state, params2, params1, result, processedParams2, processedParams1, false)
	if err != nil {
		return err
	}

	return nil
}
