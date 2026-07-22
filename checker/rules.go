package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/utils"
)

type Direction int8

const (
	DirectionRequest Direction = iota
	DirectionResponse
	DirectionNone
)

// Area is the OpenAPI Object a rule concerns, aligned with the OpenAPI spec's object model.
type Area int8

const (
	AreaSchema Area = iota
	AreaParameters
	AreaRequestBody
	AreaResponses
	AreaPaths
	AreaHeaders
	AreaSecurity
	AreaTags
	AreaComponents
	// AreaInfo and AreaServers have no rules yet: info is annotation-only
	// (title, description, version never affect the wire contract), and
	// server changes are treated as deployment metadata. They exist so the
	// area taxonomy covers every top-level document section a future rule
	// could target.
	AreaInfo
	AreaServers
	AreaNone
)

// Kind is the aspect of the API contract a rule concerns, orthogonal to Area.
type Kind int8

const (
	KindExistence Kind = iota
	KindRequiredness
	KindMutability
	KindType
	KindConstraints
	KindValues
	KindStructure
	KindLifecycle
	KindNone
)

type Action int8

const (
	ActionAdd Action = iota
	ActionRemove
	ActionChange
	ActionGeneralize
	ActionSpecialize
	ActionIncrease
	ActionDecrease
	ActionSet
	ActionNone
)

func (d Direction) String() string {
	switch d {
	case DirectionRequest:
		return "request"
	case DirectionResponse:
		return "response"
	default:
		return "none"
	}
}

func (a Area) String() string {
	switch a {
	case AreaSchema:
		return "schema"
	case AreaParameters:
		return "parameters"
	case AreaRequestBody:
		return "requestBody"
	case AreaResponses:
		return "responses"
	case AreaPaths:
		return "paths"
	case AreaHeaders:
		return "headers"
	case AreaSecurity:
		return "security"
	case AreaTags:
		return "tags"
	case AreaComponents:
		return "components"
	case AreaInfo:
		return "info"
	case AreaServers:
		return "servers"
	default:
		return "none"
	}
}

func (k Kind) String() string {
	switch k {
	case KindExistence:
		return "existence"
	case KindRequiredness:
		return "requiredness"
	case KindMutability:
		return "mutability"
	case KindType:
		return "type"
	case KindConstraints:
		return "constraints"
	case KindValues:
		return "values"
	case KindStructure:
		return "structure"
	case KindLifecycle:
		return "lifecycle"
	default:
		return "none"
	}
}

func (a Action) String() string {
	switch a {
	case ActionAdd:
		return "add"
	case ActionRemove:
		return "remove"
	case ActionChange:
		return "change"
	case ActionGeneralize:
		return "generalize"
	case ActionSpecialize:
		return "specialize"
	case ActionIncrease:
		return "increase"
	case ActionDecrease:
		return "decrease"
	case ActionSet:
		return "set"
	default:
		return "none"
	}
}

type BackwardCompatibilityRule struct {
	Id          string
	Level       Level
	Description string
	Handler     BackwardCompatibilityCheck
	Direction   Direction
	Area        Area
	Kind        Kind
	Action      Action
}

func newBackwardCompatibilityRule(id string, level Level, handler BackwardCompatibilityCheck,
	direction Direction,
	area Area,
	kind Kind,
	action Action) BackwardCompatibilityRule {
	return BackwardCompatibilityRule{
		Id:          id,
		Level:       level,
		Description: descriptionId(id),
		Handler:     handler,
		Direction:   direction,
		Area:        area,
		Kind:        kind,
		Action:      action,
	}
}

type BackwardCompatibilityRules []BackwardCompatibilityRule

func GetAllRules() BackwardCompatibilityRules {
	return BackwardCompatibilityRules{
		// Request property deprecation checks
		newBackwardCompatibilityRule(RequestPropertyDeprecatedId, INFO, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDeprecatedWithSunsetId, INFO, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDeprecatedSunsetMissingId, ERR, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDeprecatedInvalidId, ERR, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyReactivatedId, INFO, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestPropertySunsetDateTooSmallId, ERR, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange),
		// Response property deprecation checks
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedId, INFO, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedWithSunsetId, INFO, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedSunsetMissingId, ERR, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedInvalidId, ERR, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyReactivatedId, INFO, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertySunsetDateTooSmallId, ERR, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange),
		// APIAddedCheck
		newBackwardCompatibilityRule(EndpointAddedId, INFO, APIAddedCheck, DirectionNone, AreaPaths, KindExistence, ActionAdd),
		// APIComponentsSecurityUpdatedCheck
		newBackwardCompatibilityRule(APIComponentsSecurityRemovedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIComponentsSecurityAddedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(APIComponentsSecurityComponentOauthUrlUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange),
		newBackwardCompatibilityRule(APIComponentsSecurityTypeUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange),
		newBackwardCompatibilityRule(APIComponentsSecurityOauthTokenUrlUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeAddedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeRemovedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange),
		// APISecurityUpdatedCheck
		newBackwardCompatibilityRule(APISecurityRemovedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APISecurityAddedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(APISecurityScopeAddedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(APISecurityScopeRemovedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIGlobalSecurityRemovedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIGlobalSecurityAddedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(APIGlobalSecurityScopeAddedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(APIGlobalSecurityScopeRemovedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove),
		// Stability checks are run as part of CheckBackwardCompatibility.
		newBackwardCompatibilityRule(APIStabilityDecreasedId, ERR, nil, DirectionNone, AreaPaths, KindLifecycle, ActionDecrease),
		newBackwardCompatibilityRule(APIStabilityIncreasedId, INFO, nil, DirectionNone, AreaPaths, KindLifecycle, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyStabilityDecreasedId, ERR, nil, DirectionRequest, AreaSchema, KindLifecycle, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyStabilityIncreasedId, INFO, nil, DirectionRequest, AreaSchema, KindLifecycle, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyStabilityDecreasedId, ERR, nil, DirectionResponse, AreaSchema, KindLifecycle, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyStabilityIncreasedId, INFO, nil, DirectionResponse, AreaSchema, KindLifecycle, ActionIncrease),
		// APIDeprecationCheck
		newBackwardCompatibilityRule(EndpointReactivatedId, INFO, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(APIDeprecatedSunsetParseId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(APIDeprecatedSunsetMissingId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(APIInvalidStabilityLevelId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(APISunsetDateTooSmallId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(EndpointDeprecatedId, INFO, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(EndpointDeprecatedWithSunsetId, INFO, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		// RequestParameterDeprecationCheck
		newBackwardCompatibilityRule(RequestParameterReactivatedId, INFO, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestParameterDeprecatedSunsetMissingId, ERR, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestParameterSunsetDateTooSmallId, ERR, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestParameterDeprecatedId, INFO, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange),
		// APIRemovedCheck
		newBackwardCompatibilityRule(APIPathRemovedWithoutDeprecationId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIPathRemovedWithDeprecationId, INFO, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIPathSunsetParseId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(APIPathRemovedBeforeSunsetId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIRemovedWithoutDeprecationId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIRemovedWithDeprecationId, INFO, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIRemovedBeforeSunsetId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove),
		// APISunsetChangedCheck
		newBackwardCompatibilityRule(APISunsetDeletedId, ERR, APISunsetChangedCheck, DirectionNone, AreaPaths, KindLifecycle, ActionRemove),
		newBackwardCompatibilityRule(APISunsetDateChangedTooSmallId, ERR, APISunsetChangedCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange),
		// RequestParameterSunsetChangedCheck
		newBackwardCompatibilityRule(RequestParameterSunsetDeletedId, ERR, RequestParameterSunsetChangedCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(RequestParameterSunsetDateChangedTooSmallId, ERR, RequestParameterSunsetChangedCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange),
		// AddedRequiredRequestBodyCheck
		newBackwardCompatibilityRule(AddedRequiredRequestBodyId, ERR, AddedRequestBodyCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(AddedOptionalRequestBodyId, INFO, AddedRequestBodyCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd),
		// NewRequestNonPathDefaultParameterCheck
		newBackwardCompatibilityRule(NewRequiredRequestDefaultParameterToExistingPathId, ERR, NewRequestNonPathDefaultParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(NewOptionalRequestDefaultParameterToExistingPathId, INFO, NewRequestNonPathDefaultParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd),
		// NewRequestNonPathParameterCheck
		newBackwardCompatibilityRule(NewRequiredRequestParameterId, ERR, NewRequestNonPathParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(NewOptionalRequestParameterId, INFO, NewRequestNonPathParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd),
		// NewRequestPathParameterCheck
		newBackwardCompatibilityRule(NewRequestPathParameterId, ERR, NewRequestPathParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd),
		// NewRequiredRequestHeaderPropertyCheck
		newBackwardCompatibilityRule(NewRequiredRequestHeaderPropertyId, ERR, NewRequiredRequestHeaderPropertyCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd),
		// RequestBodyBecameEnumCheck
		newBackwardCompatibilityRule(RequestBodyBecameEnumId, ERR, RequestBodyBecameEnumCheck, DirectionRequest, AreaSchema, KindValues, ActionChange),
		// RequestBodyMediaTypeChangedCheck
		newBackwardCompatibilityRule(RequestBodyMediaTypeAddedId, INFO, RequestBodyMediaTypeChangedCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyMediaTypeRemovedId, ERR, RequestBodyMediaTypeChangedCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionRemove),
		// MediaTypeSchemaExistenceCheck: a schema appearing/disappearing within an existing media type (#1050).
		newBackwardCompatibilityRule(RequestBodyMediaTypeSchemaAddedId, ERR, MediaTypeSchemaExistenceCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyMediaTypeSchemaRemovedId, INFO, MediaTypeSchemaExistenceCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyMediaTypeSchemaAddedId, INFO, MediaTypeSchemaExistenceCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyMediaTypeSchemaRemovedId, WARN, MediaTypeSchemaExistenceCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove),
		// RequestBodyRemovedCheck
		newBackwardCompatibilityRule(RequestBodyRemovedId, ERR, RequestBodyRemovedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove),
		// RequestBodyRequiredUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyBecameOptionalId, INFO, RequestBodyRequiredUpdatedCheck, DirectionRequest, AreaRequestBody, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestBodyBecameRequiredId, ERR, RequestBodyRequiredUpdatedCheck, DirectionRequest, AreaRequestBody, KindRequiredness, ActionChange),
		// RequestDiscriminatorUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyDiscriminatorAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorRemovedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorPropertyNameChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingDeletedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorRemovedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorPropertyNameChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingDeletedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange),
		// RequestHeaderPropertyBecameEnumCheck
		newBackwardCompatibilityRule(RequestHeaderPropertyBecameEnumId, ERR, RequestHeaderPropertyBecameEnumCheck, DirectionRequest, AreaParameters, KindValues, ActionChange),
		// RequestHeaderPropertyBecameRequiredCheck
		newBackwardCompatibilityRule(RequestHeaderPropertyBecameRequiredId, ERR, RequestHeaderPropertyBecameRequiredCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
		// RequestParameterBecameEnumCheck
		newBackwardCompatibilityRule(RequestParameterBecameEnumId, ERR, RequestParameterBecameEnumCheck, DirectionRequest, AreaParameters, KindValues, ActionChange),
		newBackwardCompatibilityRule(RequestParameterBecameNullableId, INFO, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestParameterBecameNotNullableId, ERR, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestParameterPropertyBecameNullableId, INFO, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestParameterPropertyBecameNotNullableId, ERR, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
		// RequestParameterDefaultValueChangedCheck
		newBackwardCompatibilityRule(RequestParameterDefaultValueChangedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, AreaParameters, KindValues, ActionChange),
		newBackwardCompatibilityRule(RequestParameterDefaultValueAddedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, AreaParameters, KindValues, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterDefaultValueRemovedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove),
		// RequestParameterEnumValueUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterEnumValueAddedId, INFO, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterEnumValueRemovedId, ERR, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterPropertyEnumValueAddedId, INFO, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterPropertyEnumValueRemovedId, ERR, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove),
		// RequestParameterMaxItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxItemsIncreasedId, INFO, RequestParameterMaxItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMaxItemsDecreasedId, ERR, RequestParameterMaxItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		// RequestParameterMaxLengthSetCheck
		newBackwardCompatibilityRule(RequestParameterMaxLengthSetId, WARN, RequestParameterMaxLengthSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet),
		// RequestParameterMaxLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxLengthIncreasedId, INFO, RequestParameterMaxLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMaxLengthDecreasedId, ERR, RequestParameterMaxLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		// RequestParameterMaxSetCheck
		newBackwardCompatibilityRule(RequestParameterMaxSetId, WARN, RequestParameterMaxSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestParameterExclusiveMaxSetId, WARN, RequestParameterMaxSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet),
		// RequestParameterMaxUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxIncreasedId, INFO, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMaxDecreasedId, ERR, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestParameterExclusiveMaxIncreasedId, INFO, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterExclusiveMaxDecreasedId, ERR, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		// RequestParameterMinItemsSetCheck
		newBackwardCompatibilityRule(RequestParameterMinItemsSetId, WARN, RequestParameterMinItemsSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet),
		// RequestParameterMinItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinItemsIncreasedId, ERR, RequestParameterMinItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMinItemsDecreasedId, INFO, RequestParameterMinItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		// RequestParameterMinLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinLengthIncreasedId, ERR, RequestParameterMinLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMinLengthDecreasedId, INFO, RequestParameterMinLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		// RequestParameterMinSetCheck
		newBackwardCompatibilityRule(RequestParameterMinSetId, WARN, RequestParameterMinSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestParameterExclusiveMinSetId, WARN, RequestParameterMinSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet),
		// RequestParameterMinUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinIncreasedId, ERR, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMinDecreasedId, INFO, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestParameterExclusiveMinIncreasedId, ERR, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterExclusiveMinDecreasedId, INFO, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease),
		// RequestParameterPatternAddedOrChangedCheck
		newBackwardCompatibilityRule(RequestParameterPatternAddedId, ERR, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterPatternRemovedId, INFO, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterPatternChangedId, WARN, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionChange),
		newBackwardCompatibilityRule(RequestParameterPatternGeneralizedId, INFO, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionGeneralize),
		// RequestParameterRemovedCheck
		newBackwardCompatibilityRule(RequestParameterRemovedId, WARN, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterRemovedWithDeprecationId, INFO, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterSunsetParseId, ERR, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange),
		newBackwardCompatibilityRule(ParameterRemovedBeforeSunsetId, ERR, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindExistence, ActionRemove),
		// RequestParameterRequiredValueUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterBecomeRequiredId, ERR, RequestParameterRequiredValueUpdatedCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestParameterBecomeOptionalId, INFO, RequestParameterRequiredValueUpdatedCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
		// RequestParameterTypeChangedCheck
		newBackwardCompatibilityRule(RequestParameterTypeChangedId, ERR, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestParameterTypeGeneralizedId, INFO, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionGeneralize),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeChangedId, WARN, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeGeneralizedId, INFO, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionGeneralize),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeSpecializedId, ERR, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionSpecialize),
		// RequestParameterXExtensibleEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestParameterXExtensibleEnumValueRemovedId, ERR, RequestParameterXExtensibleEnumValueRemovedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove),
		// RequestPropertyAllOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyAllOfAddedId, ERR, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyAllOfRemovedId, WARN, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyAllOfAddedId, ERR, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyAllOfRemovedId, WARN, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyAllOfAddedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyAllOfRemovedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyAllOfAddedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyAllOfRemovedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyAnyOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyAnyOfAddedId, INFO, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyAnyOfRemovedId, ERR, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyAnyOfAddedId, INFO, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyAnyOfRemovedId, ERR, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyBecameEnumCheck
		newBackwardCompatibilityRule(RequestPropertyBecameEnumId, ERR, RequestPropertyBecameEnumCheck, DirectionRequest, AreaSchema, KindValues, ActionChange),
		// RequestPropertyBecameNotNullableCheck
		newBackwardCompatibilityRule(RequestBodyBecomeNotNullableId, ERR, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestBodyBecomeNullableId, INFO, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecomeNotNullableId, ERR, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecomeNullableId, INFO, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange),
		// RequestPropertyDefaultValueChangedCheck
		newBackwardCompatibilityRule(RequestBodyDefaultValueAddedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDefaultValueRemovedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyDefaultValueChangedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueAddedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueRemovedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueChangedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange),
		// RequestPropertyConstChangedCheck
		newBackwardCompatibilityRule(RequestBodyConstAddedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyConstRemovedId, INFO, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyConstChangedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyConstAddedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyConstRemovedId, INFO, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyConstChangedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange),
		// RequestPropertyEnumValueUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyEnumValueRemovedId, ERR, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyEnumValueRemovedId, INFO, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyEnumValueAddedId, INFO, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd),
		// RequestPropertyMaxDecreasedCheck
		newBackwardCompatibilityRule(RequestBodyMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestBodyMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxDecreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyExclusiveMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestBodyExclusiveMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyExclusiveMaxDecreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		// RequestPropertyMaxLengthSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxLengthSetId, WARN, RequestPropertyMaxLengthSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthSetId, WARN, RequestPropertyMaxLengthSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		// RequestPropertyMaxLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMaxLengthDecreasedId, ERR, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestBodyMaxLengthIncreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthDecreasedId, ERR, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxLengthDecreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthIncreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		// RequestPropertyMaxSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestBodyExclusiveMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		// RequestPropertyMinIncreasedCheck
		newBackwardCompatibilityRule(RequestBodyMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMinIncreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestBodyExclusiveMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyExclusiveMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyExclusiveMinIncreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		// RequestPropertyMinItemsIncreasedCheck
		newBackwardCompatibilityRule(RequestBodyMinItemsIncreasedId, ERR, RequestPropertyMinItemsIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMinItemsIncreasedId, ERR, RequestPropertyMinItemsIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		// RequestPropertyMinItemsSetCheck
		newBackwardCompatibilityRule(RequestBodyMinItemsSetId, WARN, RequestPropertyMinItemsSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMinItemsSetId, WARN, RequestPropertyMinItemsSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		// RequestPropertyMinLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMinLengthIncreasedId, ERR, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyMinLengthDecreasedId, INFO, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMinLengthIncreasedId, ERR, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMinLengthDecreasedId, INFO, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		// RequestPropertyMinSetCheck
		newBackwardCompatibilityRule(RequestBodyMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestBodyExclusiveMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet),
		// RequestPropertyOneOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyOneOfAddedId, INFO, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyOneOfRemovedId, ERR, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyOneOfAddedId, INFO, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyOneOfRemovedId, ERR, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyPatternUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyPatternRemovedId, INFO, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyPatternAddedId, ERR, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyPatternChangedId, WARN, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyPatternGeneralizedId, INFO, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionGeneralize),
		// RequestPropertyRequiredUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyBecameRequiredId, ERR, RequestPropertyRequiredUpdatedCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecameRequiredWithDefaultId, ERR, RequestPropertyRequiredUpdatedCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecameOptionalId, INFO, RequestPropertyRequiredUpdatedCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange),
		// RequestPropertyTypeChangedCheck
		newBackwardCompatibilityRule(RequestBodyTypeGeneralizedId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionGeneralize),
		newBackwardCompatibilityRule(RequestBodyTypeCompatibleId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestBodyTypeChangedId, ERR, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyTypeGeneralizedId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionGeneralize),
		newBackwardCompatibilityRule(RequestPropertyTypeCompatibleId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyTypeChangedId, ERR, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange),
		// RequestPropertyUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyRemovedId, WARN, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(NewRequiredRequestPropertyId, ERR, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(NewRequiredRequestPropertyWithDefaultId, ERR, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(NewOptionalRequestPropertyId, INFO, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyWrappedInOneOfId, ERR, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange),
		// RequestPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameNonWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameNonReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameNonWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameNonReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange),
		// RequestPropertyXExtensibleEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestPropertyXExtensibleEnumValueRemovedId, ERR, RequestPropertyXExtensibleEnumValueRemovedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
		// ResponseDiscriminatorUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorRemovedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorPropertyNameChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingDeletedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorRemovedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorPropertyNameChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingDeletedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange),
		// ResponseHeaderBecameOptionalCheck
		newBackwardCompatibilityRule(ResponseHeaderBecameOptionalId, ERR, ResponseHeaderBecameOptionalCheck, DirectionResponse, AreaHeaders, KindRequiredness, ActionChange),
		// ResponseHeaderRemovedCheck
		newBackwardCompatibilityRule(RequiredResponseHeaderRemovedId, ERR, ResponseHeaderRemovedCheck, DirectionResponse, AreaHeaders, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(OptionalResponseHeaderRemovedId, WARN, ResponseHeaderRemovedCheck, DirectionResponse, AreaHeaders, KindExistence, ActionRemove),
		// ResponseHeaderAddedCheck
		newBackwardCompatibilityRule(ResponseHeaderAddedId, INFO, ResponseHeaderAddedCheck, DirectionResponse, AreaHeaders, KindExistence, ActionAdd),
		// ResponseMediaTypeUpdatedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeRemovedId, ERR, ResponseMediaTypeUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseMediaTypeAddedId, INFO, ResponseMediaTypeUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd),
		// ResponseMediaTypeNameUpdatedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeNameChangedId, INFO, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionChange),
		newBackwardCompatibilityRule(ResponseMediaTypeNameGeneralizedId, ERR, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionGeneralize),
		newBackwardCompatibilityRule(ResponseMediaTypeNameSpecializedId, INFO, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionSpecialize),
		// ResponseOptionalPropertyUpdatedCheck
		newBackwardCompatibilityRule(ResponseOptionalPropertyRemovedId, WARN, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseOptionalWriteOnlyPropertyRemovedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseOptionalPropertyAddedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(ResponseOptionalWriteOnlyPropertyAddedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd),
		// ResponseOptionalPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameNonWriteOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameWriteOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameReadOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameNonReadOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		// ResponsePatternAddedOrChangedCheck
		newBackwardCompatibilityRule(ResponsePropertyPatternAddedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyPatternChangedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyPatternRemovedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove),
		// ResponsePropertyAllOfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyAllOfAddedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyAllOfRemovedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyAllOfAddedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyAllOfRemovedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyAllOfAddedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyAllOfRemovedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyAllOfAddedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyAllOfRemovedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyAnyOfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyAnyOfAddedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyAnyOfRemovedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyAnyOfAddedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyAnyOfRemovedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyBecameNullableCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameNullableId, ERR, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(ResponseBodyBecameNullableId, ERR, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyBecameNotNullableId, INFO, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(ResponseBodyBecameNotNullableId, INFO, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		// ResponsePropertyBecameOptionalCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameOptionalId, ERR, ResponsePropertyBecameOptionalCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyBecameOptionalId, INFO, ResponsePropertyBecameOptionalCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		// ResponsePropertyBecameRequiredCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameRequiredId, INFO, ResponsePropertyBecameRequiredCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyBecameRequiredId, INFO, ResponsePropertyBecameRequiredCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange),
		// ResponsePropertyDefaultValueChangedCheck
		newBackwardCompatibilityRule(ResponseBodyDefaultValueAddedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDefaultValueRemovedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyDefaultValueChangedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueAddedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueRemovedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueChangedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange),
		// ResponsePropertyConstChangedCheck
		newBackwardCompatibilityRule(ResponseBodyConstAddedId, INFO, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyConstRemovedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyConstChangedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyConstAddedId, INFO, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyConstRemovedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyConstChangedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange),
		// ResponsePropertyEnumValueAddedCheck
		newBackwardCompatibilityRule(ResponsePropertyEnumValueAddedId, WARN, ResponsePropertyEnumValueAddedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyEnumValueAddedId, INFO, ResponsePropertyEnumValueAddedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd),
		// ResponsePropertyMaxIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponseBodyExclusiveMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyExclusiveMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		// ResponsePropertyMaxLengthIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxLengthIncreasedId, ERR, ResponsePropertyMaxLengthIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyMaxLengthIncreasedId, ERR, ResponsePropertyMaxLengthIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		// ResponsePropertyMaxLengthUnsetCheck
		newBackwardCompatibilityRule(ResponseBodyMaxLengthUnsetId, ERR, ResponsePropertyMaxLengthUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyMaxLengthUnsetId, ERR, ResponsePropertyMaxLengthUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove),
		// ResponsePropertyMinDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponseBodyExclusiveMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyExclusiveMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		// ResponsePropertyMinItemsDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinItemsDecreasedId, ERR, ResponsePropertyMinItemsDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyMinItemsDecreasedId, ERR, ResponsePropertyMinItemsDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		// ResponsePropertyMinItemsUnsetCheck
		newBackwardCompatibilityRule(ResponseBodyMinItemsUnsetId, ERR, ResponsePropertyMinItemsUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyMinItemsUnsetId, ERR, ResponsePropertyMinItemsUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove),
		// ResponsePropertyMinLengthDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinLengthDecreasedId, ERR, ResponsePropertyMinLengthDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyMinLengthDecreasedId, ERR, ResponsePropertyMinLengthDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		// ResponsePropertyOneOfUpdated
		newBackwardCompatibilityRule(ResponseBodyOneOfAddedId, ERR, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyOneOfRemovedId, INFO, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyOneOfAddedId, ERR, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyOneOfRemovedId, INFO, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyTypeChangedCheck
		newBackwardCompatibilityRule(ResponseBodyTypeChangedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(ResponseBodyTypeGeneralizedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionGeneralize),
		newBackwardCompatibilityRule(ResponseBodyTypeSpecializedId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionSpecialize),
		newBackwardCompatibilityRule(ResponseBodyTypeCompatibleId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyTypeChangedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyTypeGeneralizedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionGeneralize),
		newBackwardCompatibilityRule(ResponsePropertyTypeSpecializedId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionSpecialize),
		newBackwardCompatibilityRule(ResponsePropertyTypeCompatibleId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange),
		// ResponseRequiredPropertyUpdatedCheck
		newBackwardCompatibilityRule(ResponseRequiredPropertyRemovedId, ERR, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseRequiredWriteOnlyPropertyRemovedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseRequiredPropertyAddedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(ResponseRequiredWriteOnlyPropertyAddedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyWrappedInOneOfId, ERR, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange),
		// ResponseRequiredPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameNonWriteOnlyId, WARN, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameWriteOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameReadOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameNonReadOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange),
		// ResponseSuccessStatusUpdatedCheck
		newBackwardCompatibilityRule(ResponseSuccessStatusRemovedId, ERR, ResponseSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseSuccessStatusAddedId, INFO, ResponseSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd),
		// ResponseNonSuccessStatusUpdatedCheck
		newBackwardCompatibilityRule(ResponseNonSuccessStatusRemovedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove), // optional
		newBackwardCompatibilityRule(ResponseNonSuccessStatusAddedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd),
		// APIOperationIdUpdatedCheck
		newBackwardCompatibilityRule(APIOperationIdRemovedId, INFO, APIOperationIdUpdatedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove), // optional
		newBackwardCompatibilityRule(APIOperationIdAddId, INFO, APIOperationIdUpdatedCheck, DirectionNone, AreaPaths, KindExistence, ActionAdd),
		// APITagUpdatedCheck
		newBackwardCompatibilityRule(APITagRemovedId, INFO, APITagUpdatedCheck, DirectionNone, AreaTags, KindExistence, ActionRemove), // optional
		newBackwardCompatibilityRule(APITagAddedId, INFO, APITagUpdatedCheck, DirectionNone, AreaTags, KindExistence, ActionAdd),
		// WebhookUpdatedCheck
		newBackwardCompatibilityRule(WebhookAddedId, INFO, WebhookUpdatedCheck, DirectionNone, AreaComponents, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(WebhookRemovedId, ERR, WebhookUpdatedCheck, DirectionNone, AreaComponents, KindExistence, ActionRemove),
		// APIComponentsSchemaRemovedCheck
		newBackwardCompatibilityRule(APISchemasRemovedId, INFO, APIComponentsSchemaRemovedCheck, DirectionNone, AreaComponents, KindExistence, ActionRemove), // optional
		// ResponseParameterEnumValueRemovedCheck
		newBackwardCompatibilityRule(ResponsePropertyEnumValueRemovedId, INFO, ResponseParameterEnumValueRemovedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove), // optional
		// ResponseMediaTypeEnumValueRemovedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeEnumValueRemovedId, INFO, ResponseMediaTypeEnumValueRemovedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove), // optional
		// RequestBodyEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestBodyEnumValueRemovedId, INFO, RequestBodyEnumValueRemovedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove), // optional
		// RequestPropertyListOfTypesChangedCheck
		newBackwardCompatibilityRule(RequestBodyListOfTypesWidenedId, INFO, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyListOfTypesNarrowedId, ERR, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyListOfTypesWidenedId, INFO, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyListOfTypesNarrowedId, ERR, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionRemove),
		// ResponsePropertyListOfTypesChangedCheck
		newBackwardCompatibilityRule(ResponseBodyListOfTypesWidenedId, ERR, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyListOfTypesNarrowedId, INFO, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyListOfTypesWidenedId, ERR, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyListOfTypesNarrowedId, INFO, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionRemove),
		// RequestParameterListOfTypesChangedCheck
		newBackwardCompatibilityRule(RequestParameterListOfTypesWidenedId, INFO, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterListOfTypesNarrowedId, ERR, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterPropertyListOfTypesWidenedId, INFO, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterPropertyListOfTypesNarrowedId, ERR, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionRemove),
		// RequestPropertyPrefixItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyPrefixItemsAddedId, INFO, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyPrefixItemsRemovedId, ERR, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyPrefixItemsAddedId, INFO, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyPrefixItemsRemovedId, ERR, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyPrefixItemsUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyPrefixItemsAddedId, ERR, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyPrefixItemsRemovedId, INFO, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyPrefixItemsAddedId, ERR, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyPrefixItemsRemovedId, INFO, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyIfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyIfAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyIfRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyThenAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyThenRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyElseAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyElseRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyIfAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyIfRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyThenAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyThenRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyElseAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyElseRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyIfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyIfAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyIfRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyThenAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyThenRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyElseAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyElseRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyIfAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyIfRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyThenAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyThenRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyElseAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyElseRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyContainsUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyContainsAddedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyContainsRemovedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyMinContainsIncreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyMinContainsDecreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestBodyMaxContainsIncreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyMaxContainsDecreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyContainsAddedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyContainsRemovedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyMinContainsIncreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMinContainsDecreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMaxContainsIncreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMaxContainsDecreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease),
		// ResponsePropertyContainsUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyContainsAddedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyContainsRemovedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyMinContainsIncreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponseBodyMinContainsDecreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponseBodyMaxContainsIncreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponseBodyMaxContainsDecreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyContainsAddedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyContainsRemovedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyMinContainsIncreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyMinContainsDecreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyMaxContainsIncreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyMaxContainsDecreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease),
		// RequestPropertyDependentRequiredChangedCheck
		newBackwardCompatibilityRule(RequestBodyDependentRequiredAddedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDependentRequiredRemovedId, INFO, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyDependentRequiredChangedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDependentRequiredAddedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDependentRequiredRemovedId, INFO, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDependentRequiredChangedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange),
		// ResponsePropertyDependentRequiredChangedCheck
		newBackwardCompatibilityRule(ResponseBodyDependentRequiredAddedId, INFO, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDependentRequiredRemovedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyDependentRequiredChangedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDependentRequiredAddedId, INFO, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDependentRequiredRemovedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDependentRequiredChangedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange),
		// RequestPropertyDependentSchemasUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyDependentSchemaAddedId, ERR, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDependentSchemaRemovedId, INFO, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDependentSchemaAddedId, ERR, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDependentSchemaRemovedId, INFO, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyDependentSchemasUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyDependentSchemaAddedId, INFO, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDependentSchemaRemovedId, ERR, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDependentSchemaAddedId, INFO, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDependentSchemaRemovedId, ERR, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyPatternPropertiesUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyPatternPropertyAddedId, ERR, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyPatternPropertyRemovedId, INFO, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyPatternPropertyAddedId, ERR, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyPatternPropertyRemovedId, INFO, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove),
		// ResponsePropertyPatternPropertiesUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyPatternPropertyAddedId, INFO, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyPatternPropertyRemovedId, ERR, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyPatternPropertyAddedId, INFO, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyPatternPropertyRemovedId, ERR, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove),
		// RequestPropertyPropertyNamesUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyPropertyNamesAddedId, ERR, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyPropertyNamesRemovedId, INFO, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyPropertyNamesAddedId, ERR, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyPropertyNamesRemovedId, INFO, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyPropertyNamesUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyPropertyNamesAddedId, INFO, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyPropertyNamesRemovedId, ERR, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyPropertyNamesAddedId, INFO, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyPropertyNamesRemovedId, ERR, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyUnevaluatedUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyUnevaluatedItemsAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyUnevaluatedItemsRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyUnevaluatedPropertiesAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyUnevaluatedPropertiesRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedItemsAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedItemsRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedPropertiesAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedPropertiesRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove),
		// ResponsePropertyUnevaluatedUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedItemsAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedItemsRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedPropertiesAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedPropertiesRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedItemsAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedItemsRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedPropertiesAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedPropertiesRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove),
		// RequestPropertyContentUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyContentSchemaAddedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyContentSchemaRemovedId, INFO, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyContentMediaTypeChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaRequestBody, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestBodyContentEncodingChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaRequestBody, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyContentSchemaAddedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyContentSchemaRemovedId, INFO, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyContentMediaTypeChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyContentEncodingChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindType, ActionChange),
		// ResponsePropertyContentUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyContentSchemaAddedId, INFO, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyContentSchemaRemovedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyContentMediaTypeChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionChange),
		newBackwardCompatibilityRule(ResponseBodyContentEncodingChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyContentSchemaAddedId, INFO, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyContentSchemaRemovedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyContentMediaTypeChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindType, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyContentEncodingChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindType, ActionChange),
	}
}

func GetOptionalRules() BackwardCompatibilityRules {
	return BackwardCompatibilityRules{
		newBackwardCompatibilityRule(ResponseNonSuccessStatusRemovedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APIOperationIdRemovedId, INFO, APIOperationIdUpdatedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APITagRemovedId, INFO, APITagUpdatedCheck, DirectionNone, AreaTags, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(APISchemasRemovedId, INFO, APIComponentsSchemaRemovedCheck, DirectionNone, AreaComponents, KindExistence, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyEnumValueRemovedId, INFO, ResponseParameterEnumValueRemovedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(ResponseMediaTypeEnumValueRemovedId, INFO, ResponseMediaTypeEnumValueRemovedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyEnumValueRemovedId, INFO, RequestBodyEnumValueRemovedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove),
	}
}

// GetCheckLevels gets levels for all backward compatibility checks
func GetCheckLevels() map[string]Level {
	return rulesToLevels(GetAllRules())
}

// GetAllChecks gets all backward compatibility checks
func GetAllChecks() BackwardCompatibilityChecks {
	return rulesToChecks(GetAllRules())
}

// rulesToChecks return a unique list of checks from a list of rules
func rulesToChecks(rules BackwardCompatibilityRules) BackwardCompatibilityChecks {
	result := BackwardCompatibilityChecks{}
	m := utils.StringSet{}
	for _, rule := range rules {
		// functions are not comparable, so we convert them to strings
		pStr := fmt.Sprintf("%v", rule.Handler)
		if !m.Contains(pStr) {
			m.Add(pStr)
			result = append(result, rule.Handler)
		}
	}
	return result
}

func GetOptionalRuleIds() []string {
	return rulesToIIs(GetOptionalRules())
}

func GetAllRuleIds() []string {
	return rulesToIIs(GetAllRules())
}

// rulesToLevels return a map of check IDs to levels
func rulesToLevels(rules BackwardCompatibilityRules) map[string]Level {
	result := map[string]Level{}
	for _, rule := range rules {
		result[rule.Id] = rule.Level
	}
	return result
}

func rulesToIIs(rules BackwardCompatibilityRules) []string {
	result := []string{}
	for _, rule := range rules {
		result = append(result, rule.Id)
	}
	return result
}
