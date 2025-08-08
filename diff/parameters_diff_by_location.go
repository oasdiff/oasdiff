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

	// Track which exploded parameters have been matched to avoid multiple matches
	explodedParamsMatched := make(map[*openapi3.Parameter]bool)

	for _, paramRef1 := range params1 {
		param1, err := derefParam(paramRef1)
		if err != nil {
			return nil, err
		}

		param2, err := findParam(param1, params2, pathParamsMap)
		if err != nil {
			return nil, err
		}

		if param2 != nil {
			// Check if this is an exploded parameter match
			isExplodedMatch := false
			if isExplodedObjectParam(param2) {
				isEquivalent, err := isParamInExplodedObject(param1, param2)
				if err != nil {
					return nil, err
				}
				if isEquivalent {
					isExplodedMatch = true
					// Mark this exploded parameter as matched
					explodedParamsMatched[param2] = true
				}
			}

			// Skip diff generation for exploded matches to avoid type change warnings
			if !isExplodedMatch {
				diff, err := getParameterDiff(config, state, param1, param2)
				if err != nil {
					return nil, err
				}

				if !diff.Empty() {
					result.addModifiedParam(param1, diff)
				}
			}
		} else {
			result.addDeletedParam(param1)
		}
	}

	pathParamsMapInversed := pathParamsMap.Inverse()
	for _, paramRef2 := range params2 {
		param2, err := derefParam(paramRef2)
		if err != nil {
			return nil, err
		}

		param, err := findParam(param2, params1, pathParamsMapInversed)
		if err != nil {
			return nil, err
		}
		if param == nil {
			// Don't flag exploded parameters as added if they were matched to simple parameters
			if explodedParamsMatched[param2] {
				continue
			}
			result.addAddedParam(param2)
		}
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
