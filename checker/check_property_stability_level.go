package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyStabilityDecreasedId  = "request-property-stability-decreased"
	RequestPropertyStabilityIncreasedId  = "request-property-stability-increased"
	ResponsePropertyStabilityDecreasedId = "response-property-stability-decreased"
	ResponsePropertyStabilityIncreasedId = "response-property-stability-increased"
)

// RequestPropertyStabilityUpdatedCheck detects request properties where x-stability-level changed.
// Only emits changes when the property's base stability meets the configured threshold.
func RequestPropertyStabilityUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if config == nil {
		return result
	}
	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			checkPropertyStabilityChange(p,
				RequestPropertyStabilityDecreasedId, RequestPropertyStabilityIncreasedId, &result)
		})
	})
	return result
}

// ResponsePropertyStabilityUpdatedCheck detects response properties where x-stability-level changed.
// Only emits changes when the property's base stability meets the configured threshold.
func ResponsePropertyStabilityUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if config == nil {
		return result
	}
	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			checkPropertyStabilityChange(p,
				ResponsePropertyStabilityDecreasedId, ResponsePropertyStabilityIncreasedId, &result)
		})
	})
	return result
}

func checkPropertyStabilityChange(p propertyInfo, decreasedId string, increasedId string, result *Changes) {
	if p.propertyDiff.ExtensionsDiff == nil {
		return
	}

	op := p.operationItem.Revision

	// An unparseable x-stability-level on a property is reported here: unlike
	// the endpoint case, checkInvalidStabilityLevels does not descend into
	// property schemas, so there is no other backstop for it.
	baseStability, err := getStabilityLevel(p.propertyDiff.Base.Extensions)
	if err != nil {
		baseSource := stabilityFieldSource(p.operationsSources, op, p.propertyDiff.Base.Origin)
		*result = append(*result, getAPIInvalidStabilityLevel(p.config, op, p.operationsSources, p.method, p.path, err).WithSources(baseSource, nil))
		return
	}
	revisionStability, err := getStabilityLevel(p.propertyDiff.Revision.Extensions)
	if err != nil {
		revisionSource := stabilityFieldSource(p.operationsSources, op, p.propertyDiff.Revision.Origin)
		*result = append(*result, getAPIInvalidStabilityLevel(p.config, op, p.operationsSources, p.method, p.path, err).WithSources(nil, revisionSource))
		return
	}

	baseLabel := normalizedStability(baseStability)
	revisionLabel := normalizedStability(revisionStability)
	baseLevel := ParseStabilityLevel(baseLabel)
	revisionLevel := ParseStabilityLevel(revisionLabel)

	if baseLevel == revisionLevel {
		return
	}

	// Gate on the base stability (the level the property is leaving), consistent
	// with the endpoint-level check, so a destabilization from a tracked level is
	// reported rather than dropped.
	if !p.config.StabilityLevel.IsIncluded(baseLabel) {
		return
	}

	changeId := increasedId
	if revisionLevel < baseLevel {
		changeId = decreasedId
	}

	baseSource := stabilityFieldSource(p.operationsSources, op, p.propertyDiff.Base.Origin)
	revisionSource := stabilityFieldSource(p.operationsSources, op, p.propertyDiff.Revision.Origin)

	*result = append(*result, p.newChange(
		changeId,
		[]any{propertyFullName(p.propertyPath, p.propertyName), baseLabel, revisionLabel},
		"",
	).WithSources(baseSource, revisionSource))
}
