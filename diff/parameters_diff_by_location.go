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
		return isSemanticEquivalent(param1, param2)
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

// Patch applies the patch to parameters
func (diff *ParametersDiffByLocation) Patch(parameters openapi3.Parameters) error {

	if diff.Empty() {
		return nil
	}

	for location, paramDiffs := range diff.Modified {
		for name, parameterDiff := range paramDiffs {
			err := parameterDiff.Patch(parameters.GetByInAndName(location, name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// isSemanticEquivalent checks if two parameters are semantically equivalent,
// considering the case where separate parameters are consolidated into an exploded object parameter
func isSemanticEquivalent(param1, param2 *openapi3.Parameter) (bool, error) {
	// Check if param1 is a simple parameter and param2 is an exploded object
	if isExplodedObjectParam(param2) {
		return isParamInExplodedObject(param1, param2)
	}

	// Check if param2 is a simple parameter and param1 is an exploded object
	if isExplodedObjectParam(param1) {
		return isParamInExplodedObject(param2, param1)
	}

	return false, nil
}

// isExplodedObjectParam checks if a parameter is an exploded object parameter
// (style=form, explode=true, object schema)
func isExplodedObjectParam(param *openapi3.Parameter) bool {
	if param == nil || param.Schema == nil || param.Schema.Value == nil {
		return false
	}

	// Must be explode=true and style=form (or default)
	explode := param.Explode != nil && *param.Explode
	isFormStyle := param.Style == "" || param.Style == "form" // form is default for query params

	// Must have object schema with properties
	schema := param.Schema.Value
	isObjectWithProps := schema.Type != nil && len(*schema.Type) == 1 && (*schema.Type)[0] == "object" && len(schema.Properties) > 0

	return explode && isFormStyle && isObjectWithProps
}

// isParamInExplodedObject checks if a simple parameter corresponds to a property
// in an exploded object parameter
func isParamInExplodedObject(simpleParam, explodedParam *openapi3.Parameter) (bool, error) {
	if !isExplodedObjectParam(explodedParam) {
		return false, nil
	}

	schema := explodedParam.Schema.Value
	if schema == nil || schema.Properties == nil {
		return false, nil
	}

	// Check if the simple parameter's name matches a property in the object
	_, exists := schema.Properties[simpleParam.Name]
	return exists, nil
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
		Name:     simpleParam.Name,      // Must match for comparison
		In:       simpleParam.In,        // Must match for comparison
		Schema:   propertySchemaRef,     // The key thing we're comparing
		Required: explodedParam.Required, // Use exploded param's required status
		Style:    explodedParam.Style,   // Use exploded param's style
		Explode:  explodedParam.Explode, // Use exploded param's explode

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

// handleExplodedParameterMatching handles bidirectional equivalence between exploded parameters and simple parameters
func handleExplodedParameterMatching(config *Config, state *state, params1, params2 openapi3.Parameters, result *ParametersDiffByLocation, processedParams1, processedParams2 map[*openapi3.Parameter]bool) error {
	
	// Handle case 1: Simple parameters (params1) -> Exploded parameter (params2)
	err := handleParameterMatching(config, state, params1, params2, result, processedParams1, processedParams2, true)
	if err != nil {
		return err
	}
	
	// Handle case 2: Exploded parameter (params1) -> Simple parameters (params2)  
	err = handleParameterMatching(config, state, params1, params2, result, processedParams1, processedParams2, false)
	if err != nil {
		return err
	}
	
	return nil
}

// handleParameterMatching handles bidirectional parameter matching between exploded and simple parameters.
// The direction parameter controls which direction to process:
// - true: simple → exploded (find exploded parameters in params2 that match simple parameters in params1)
// - false: exploded → simple (find exploded parameters in params1 that match simple parameters in params2)
func handleParameterMatching(config *Config, state *state, params1, params2 openapi3.Parameters, result *ParametersDiffByLocation, processedParams1, processedParams2 map[*openapi3.Parameter]bool, simpleToExploded bool) error {
	
	var explodedParams, simpleParams openapi3.Parameters
	var explodedProcessed, simpleProcessed map[*openapi3.Parameter]bool
	
	if simpleToExploded {
		// Simple → Exploded: look for exploded params in params2, simple params in params1
		explodedParams = params2
		simpleParams = params1
		explodedProcessed = processedParams2
		simpleProcessed = processedParams1
	} else {
		// Exploded → Simple: look for exploded params in params1, simple params in params2
		explodedParams = params1
		simpleParams = params2
		explodedProcessed = processedParams1
		simpleProcessed = processedParams2
	}
	
	// Find exploded parameters
	for _, explodedParamRef := range explodedParams {
		explodedParam, err := derefParam(explodedParamRef)
		if err != nil {
			return err
		}
		
		if !isExplodedObjectParam(explodedParam) || explodedProcessed[explodedParam] {
			continue
		}
		
		// Find all simple parameters that match properties of this exploded parameter
		var matchingParams []*openapi3.Parameter
		for _, simpleParamRef := range simpleParams {
			simpleParam, err := derefParam(simpleParamRef)
			if err != nil {
				return err
			}
			
			if simpleProcessed[simpleParam] {
				continue
			}
			
			equivalent, err := isParamInExplodedObject(simpleParam, explodedParam)
			if err != nil {
				return err
			}
			
			if equivalent {
				matchingParams = append(matchingParams, simpleParam)
			}
		}
		
		// If we found matching simple parameters, process them
		if len(matchingParams) > 0 {
			for _, simpleParam := range matchingParams {
				// Generate property-level diff in the appropriate direction
				propertyDiff, err := getPropertyDiff(config, state, simpleParam, explodedParam, simpleToExploded)
				if err != nil {
					return err
				}
				
				if !propertyDiff.Empty() {
					// Use the simple parameter name for reporting
					result.addModifiedParam(simpleParam, propertyDiff)
				}
				
				// Mark simple parameter as processed
				simpleProcessed[simpleParam] = true
			}
			// Mark exploded parameter as processed
			explodedProcessed[explodedParam] = true
		}
	}
	
	return nil
}
