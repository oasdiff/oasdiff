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

type Location int8

const (
	LocationBody Location = iota
	LocationParameters
	LocationProperties
	LocationHeaders
	LocationSecurity
	LocationComponents
	LocationNone
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

type BackwardCompatibilityRule struct {
	Id          string
	Level       Level
	Description string
	Handler     BackwardCompatibilityCheck
	Direction   Direction
	Location    Location
	Action      Action
}

func newBackwardCompatibilityRule(id string, level Level, handler BackwardCompatibilityCheck,
	direction Direction,
	location Location,
	action Action) BackwardCompatibilityRule {
	return BackwardCompatibilityRule{
		Id:          id,
		Level:       level,
		Description: descriptionId(id),
		Handler:     handler,
		Direction:   direction,
		Location:    location,
		Action:      action,
	}
}

type BackwardCompatibilityRules []BackwardCompatibilityRule

func GetAllRules() BackwardCompatibilityRules {
	return BackwardCompatibilityRules{
		// APIAddedCheck
		newBackwardCompatibilityRule(EndpointAddedId, INFO, APIAddedCheck, DirectionNone, LocationNone, ActionAdd),
		// APIComponentsSecurityUpdatedCheck
		newBackwardCompatibilityRule(APIComponentsSecurityRemovedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionRemove),
		newBackwardCompatibilityRule(APIComponentsSecurityAddedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionAdd),
		newBackwardCompatibilityRule(APIComponentsSecurityComponentOauthUrlUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionChange),
		newBackwardCompatibilityRule(APIComponentsSecurityTypeUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionChange),
		newBackwardCompatibilityRule(APIComponentsSecurityOauthTokenUrlUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionChange),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeAddedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionAdd),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeRemovedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionRemove),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, LocationComponents, ActionChange),
		// APISecurityUpdatedCheck
		newBackwardCompatibilityRule(APISecurityRemovedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionRemove),
		newBackwardCompatibilityRule(APISecurityAddedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionAdd),
		newBackwardCompatibilityRule(APISecurityScopeAddedId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionAdd),
		newBackwardCompatibilityRule(APISecurityScopeRemovedId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionRemove),
		newBackwardCompatibilityRule(APIGlobalSecurityRemovedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionRemove),
		newBackwardCompatibilityRule(APIGlobalSecurityAddedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionAdd),
		newBackwardCompatibilityRule(APIGlobalSecurityScopeAddedId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionAdd),
		newBackwardCompatibilityRule(APIGlobalSecurityScopeRemovedId, INFO, APISecurityUpdatedCheck, DirectionNone, LocationSecurity, ActionRemove),
		// Stability Descreased Check is run as part of CheckBackwardCompatibility
		newBackwardCompatibilityRule(APIStabilityDecreasedId, ERR, nil, DirectionNone, LocationNone, ActionDecrease),
		// APIDeprecationCheck
		newBackwardCompatibilityRule(EndpointReactivatedId, INFO, APIDeprecationCheck, DirectionNone, LocationNone, ActionChange),
		newBackwardCompatibilityRule(APIDeprecatedSunsetParseId, ERR, APIDeprecationCheck, DirectionNone, LocationNone, ActionChange),
		newBackwardCompatibilityRule(APIDeprecatedSunsetMissingId, ERR, APIDeprecationCheck, DirectionNone, LocationNone, ActionChange),
		newBackwardCompatibilityRule(APIInvalidStabilityLevelId, ERR, APIDeprecationCheck, DirectionNone, LocationNone, ActionChange),
		newBackwardCompatibilityRule(APISunsetDateTooSmallId, ERR, APIDeprecationCheck, DirectionNone, LocationNone, ActionChange),
		newBackwardCompatibilityRule(EndpointDeprecatedId, INFO, APIDeprecationCheck, DirectionNone, LocationNone, ActionChange),
		// RequestParameterDeprecationCheck
		newBackwardCompatibilityRule(RequestParameterReactivatedId, INFO, RequestParameterDeprecationCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterDeprecatedSunsetMissingId, ERR, RequestParameterDeprecationCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterSunsetDateTooSmallId, ERR, RequestParameterDeprecationCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterDeprecatedId, INFO, RequestParameterDeprecationCheck, DirectionRequest, LocationParameters, ActionChange),
		// APIRemovedCheck
		newBackwardCompatibilityRule(APIPathRemovedWithoutDeprecationId, ERR, APIRemovedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APIPathRemovedWithDeprecationId, INFO, APIRemovedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APIPathSunsetParseId, ERR, APIRemovedCheck, DirectionNone, LocationNone, ActionChange),
		newBackwardCompatibilityRule(APIPathRemovedBeforeSunsetId, ERR, APIRemovedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APIRemovedWithoutDeprecationId, ERR, APIRemovedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APIRemovedWithDeprecationId, INFO, APIRemovedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APIRemovedBeforeSunsetId, ERR, APIRemovedCheck, DirectionNone, LocationNone, ActionRemove),
		// APISunsetChangedCheck
		newBackwardCompatibilityRule(APISunsetDeletedId, ERR, APISunsetChangedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APISunsetDateChangedTooSmallId, ERR, APISunsetChangedCheck, DirectionNone, LocationNone, ActionChange),
		// RequestParameterSunsetChangedCheck
		newBackwardCompatibilityRule(RequestParameterSunsetDeletedId, ERR, RequestParameterSunsetChangedCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterSunsetDateChangedTooSmallId, ERR, RequestParameterSunsetChangedCheck, DirectionRequest, LocationParameters, ActionChange),
		// AddedRequiredRequestBodyCheck
		newBackwardCompatibilityRule(AddedRequiredRequestBodyId, ERR, AddedRequestBodyCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(AddedOptionalRequestBodyId, INFO, AddedRequestBodyCheck, DirectionRequest, LocationBody, ActionAdd),
		// NewRequestNonPathDefaultParameterCheck
		newBackwardCompatibilityRule(NewRequiredRequestDefaultParameterToExistingPathId, ERR, NewRequestNonPathDefaultParameterCheck, DirectionRequest, LocationParameters, ActionAdd),
		newBackwardCompatibilityRule(NewOptionalRequestDefaultParameterToExistingPathId, INFO, NewRequestNonPathDefaultParameterCheck, DirectionRequest, LocationParameters, ActionAdd),
		// NewRequestNonPathParameterCheck
		newBackwardCompatibilityRule(NewRequiredRequestParameterId, ERR, NewRequestNonPathParameterCheck, DirectionRequest, LocationParameters, ActionAdd),
		newBackwardCompatibilityRule(NewOptionalRequestParameterId, INFO, NewRequestNonPathParameterCheck, DirectionRequest, LocationParameters, ActionAdd),
		// NewRequestPathParameterCheck
		newBackwardCompatibilityRule(NewRequestPathParameterId, ERR, NewRequestPathParameterCheck, DirectionRequest, LocationParameters, ActionAdd),
		// NewRequiredRequestHeaderPropertyCheck
		newBackwardCompatibilityRule(NewRequiredRequestHeaderPropertyId, ERR, NewRequiredRequestHeaderPropertyCheck, DirectionRequest, LocationProperties, ActionAdd),
		// RequestBodyBecameEnumCheck
		newBackwardCompatibilityRule(RequestBodyBecameEnumId, ERR, RequestBodyBecameEnumCheck, DirectionRequest, LocationBody, ActionChange),
		// RequestBodyMediaTypeChangedCheck
		newBackwardCompatibilityRule(RequestBodyMediaTypeAddedId, INFO, RequestBodyMediaTypeChangedCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyMediaTypeRemovedId, ERR, RequestBodyMediaTypeChangedCheck, DirectionRequest, LocationBody, ActionRemove),
		// RequestBodyRemovedCheck
		newBackwardCompatibilityRule(RequestBodyRemovedId, ERR, RequestBodyRemovedCheck, DirectionRequest, LocationBody, ActionRemove),
		// RequestBodyRequiredUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyBecameOptionalId, INFO, RequestBodyRequiredUpdatedCheck, DirectionRequest, LocationBody, ActionChange),
		newBackwardCompatibilityRule(RequestBodyBecameRequiredId, ERR, RequestBodyRequiredUpdatedCheck, DirectionRequest, LocationBody, ActionChange),
		// RequestDiscriminatorUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyDiscriminatorAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorRemovedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorPropertyNameChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationBody, ActionChange),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingDeletedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationBody, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorRemovedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorPropertyNameChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingDeletedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestHeaderPropertyBecameEnumCheck
		newBackwardCompatibilityRule(RequestHeaderPropertyBecameEnumId, ERR, RequestHeaderPropertyBecameEnumCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestHeaderPropertyBecameRequiredCheck
		newBackwardCompatibilityRule(RequestHeaderPropertyBecameRequiredId, ERR, RequestHeaderPropertyBecameRequiredCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestParameterBecameEnumCheck
		newBackwardCompatibilityRule(RequestParameterBecameEnumId, ERR, RequestParameterBecameEnumCheck, DirectionRequest, LocationParameters, ActionChange),
		// RequestParameterDefaultValueChangedCheck
		newBackwardCompatibilityRule(RequestParameterDefaultValueChangedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterDefaultValueAddedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, LocationParameters, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterDefaultValueRemovedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, LocationParameters, ActionRemove),
		// RequestParameterEnumValueUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterEnumValueAddedId, INFO, RequestParameterEnumValueUpdatedCheck, DirectionRequest, LocationParameters, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterEnumValueRemovedId, ERR, RequestParameterEnumValueUpdatedCheck, DirectionRequest, LocationParameters, ActionRemove),
		// RequestParameterMaxItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxItemsIncreasedId, INFO, RequestParameterMaxItemsUpdatedCheck, DirectionRequest, LocationParameters, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMaxItemsDecreasedId, ERR, RequestParameterMaxItemsUpdatedCheck, DirectionRequest, LocationParameters, ActionDecrease),
		// RequestParameterMaxLengthSetCheck
		newBackwardCompatibilityRule(RequestParameterMaxLengthSetId, WARN, RequestParameterMaxLengthSetCheck, DirectionRequest, LocationParameters, ActionSet),
		// RequestParameterMaxLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxLengthIncreasedId, INFO, RequestParameterMaxLengthUpdatedCheck, DirectionRequest, LocationParameters, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMaxLengthDecreasedId, ERR, RequestParameterMaxLengthUpdatedCheck, DirectionRequest, LocationParameters, ActionDecrease),
		// RequestParameterMaxSetCheck
		newBackwardCompatibilityRule(RequestParameterMaxSetId, WARN, RequestParameterMaxSetCheck, DirectionRequest, LocationParameters, ActionSet),
		// RequestParameterMaxUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxIncreasedId, INFO, RequestParameterMaxUpdatedCheck, DirectionRequest, LocationParameters, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMaxDecreasedId, ERR, RequestParameterMaxUpdatedCheck, DirectionRequest, LocationParameters, ActionDecrease),
		// RequestParameterMinItemsSetCheck
		newBackwardCompatibilityRule(RequestParameterMinItemsSetId, WARN, RequestParameterMinItemsSetCheck, DirectionRequest, LocationParameters, ActionSet),
		// RequestParameterMinItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinItemsIncreasedId, ERR, RequestParameterMinItemsUpdatedCheck, DirectionRequest, LocationParameters, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMinItemsDecreasedId, INFO, RequestParameterMinItemsUpdatedCheck, DirectionRequest, LocationParameters, ActionDecrease),
		// RequestParameterMinLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinLengthIncreasedId, ERR, RequestParameterMinLengthUpdatedCheck, DirectionRequest, LocationParameters, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMinLengthDecreasedId, INFO, RequestParameterMinLengthUpdatedCheck, DirectionRequest, LocationParameters, ActionDecrease),
		// RequestParameterMinSetCheck
		newBackwardCompatibilityRule(RequestParameterMinSetId, WARN, RequestParameterMinSetCheck, DirectionRequest, LocationParameters, ActionSet),
		// RequestParameterMinUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinIncreasedId, ERR, RequestParameterMinUpdatedCheck, DirectionRequest, LocationParameters, ActionIncrease),
		newBackwardCompatibilityRule(RequestParameterMinDecreasedId, INFO, RequestParameterMinUpdatedCheck, DirectionRequest, LocationParameters, ActionDecrease),
		// RequestParameterPatternAddedOrChangedCheck
		newBackwardCompatibilityRule(RequestParameterPatternAddedId, WARN, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, LocationParameters, ActionAdd),
		newBackwardCompatibilityRule(RequestParameterPatternRemovedId, INFO, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, LocationParameters, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterPatternChangedId, WARN, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterPatternGeneralizedId, INFO, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, LocationParameters, ActionGeneralize),
		// RequestParameterRemovedCheck
		newBackwardCompatibilityRule(RequestParameterRemovedId, WARN, RequestParameterRemovedCheck, DirectionRequest, LocationParameters, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterRemovedWithDeprecationId, INFO, RequestParameterRemovedCheck, DirectionRequest, LocationParameters, ActionRemove),
		newBackwardCompatibilityRule(RequestParameterSunsetParseId, ERR, RequestParameterRemovedCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(ParameterRemovedBeforeSunsetId, ERR, RequestParameterRemovedCheck, DirectionRequest, LocationParameters, ActionRemove),
		// RequestParameterRequiredValueUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterBecomeRequiredId, ERR, RequestParameterRequiredValueUpdatedCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterBecomeOptionalId, INFO, RequestParameterRequiredValueUpdatedCheck, DirectionRequest, LocationParameters, ActionChange),
		// RequestParameterTypeChangedCheck
		newBackwardCompatibilityRule(RequestParameterTypeChangedId, ERR, RequestParameterTypeChangedCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterTypeGeneralizedId, INFO, RequestParameterTypeChangedCheck, DirectionRequest, LocationParameters, ActionGeneralize),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeChangedId, WARN, RequestParameterTypeChangedCheck, DirectionRequest, LocationParameters, ActionChange),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeGeneralizedId, INFO, RequestParameterTypeChangedCheck, DirectionRequest, LocationParameters, ActionGeneralize),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeSpecializedId, ERR, RequestParameterTypeChangedCheck, DirectionRequest, LocationParameters, ActionSpecialize),
		// RequestParameterXExtensibleEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestParameterXExtensibleEnumValueRemovedId, ERR, RequestParameterXExtensibleEnumValueRemovedCheck, DirectionRequest, LocationParameters, ActionRemove),
		// RequestPropertyAllOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyAllOfAddedId, ERR, RequestPropertyAllOfUpdatedCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyAllOfRemovedId, WARN, RequestPropertyAllOfUpdatedCheck, DirectionRequest, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyAllOfAddedId, ERR, RequestPropertyAllOfUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyAllOfRemovedId, WARN, RequestPropertyAllOfUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		// RequestPropertyAnyOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyAnyOfAddedId, INFO, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyAnyOfRemovedId, ERR, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyAnyOfAddedId, INFO, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyAnyOfRemovedId, ERR, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		// RequestPropertyBecameEnumCheck
		newBackwardCompatibilityRule(RequestPropertyBecameEnumId, ERR, RequestPropertyBecameEnumCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestPropertyBecameNotNullableCheck
		newBackwardCompatibilityRule(RequestBodyBecomeNotNullableId, ERR, RequestPropertyBecameNotNullableCheck, DirectionRequest, LocationBody, ActionChange),
		newBackwardCompatibilityRule(RequestBodyBecomeNullableId, INFO, RequestPropertyBecameNotNullableCheck, DirectionRequest, LocationBody, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecomeNotNullableId, ERR, RequestPropertyBecameNotNullableCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecomeNullableId, INFO, RequestPropertyBecameNotNullableCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestPropertyDefaultValueChangedCheck
		newBackwardCompatibilityRule(RequestBodyDefaultValueAddedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyDefaultValueRemovedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyDefaultValueChangedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, LocationBody, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueAddedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueRemovedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueChangedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestPropertyEnumValueUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyEnumValueRemovedId, ERR, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyEnumValueRemovedId, INFO, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyEnumValueAddedId, INFO, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		// RequestPropertyMaxDecreasedCheck
		newBackwardCompatibilityRule(RequestBodyMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, LocationBody, ActionDecrease),
		newBackwardCompatibilityRule(RequestBodyMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, LocationBody, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, LocationProperties, ActionDecrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxDecreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, LocationProperties, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, LocationProperties, ActionIncrease),
		// RequestPropertyMaxLengthSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxLengthSetId, WARN, RequestPropertyMaxLengthSetCheck, DirectionRequest, LocationBody, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthSetId, WARN, RequestPropertyMaxLengthSetCheck, DirectionRequest, LocationProperties, ActionSet),
		// RequestPropertyMaxLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMaxLengthDecreasedId, ERR, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, LocationBody, ActionDecrease),
		newBackwardCompatibilityRule(RequestBodyMaxLengthIncreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, LocationBody, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthDecreasedId, ERR, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, LocationProperties, ActionDecrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxLengthDecreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, LocationProperties, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthIncreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, LocationProperties, ActionIncrease),
		// RequestPropertyMaxSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, LocationBody, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, LocationProperties, ActionSet),
		// RequestPropertyMinIncreasedCheck
		newBackwardCompatibilityRule(RequestBodyMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, LocationBody, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, LocationBody, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, LocationProperties, ActionIncrease),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMinIncreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, LocationProperties, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, LocationProperties, ActionDecrease),
		// RequestPropertyMinItemsIncreasedCheck
		newBackwardCompatibilityRule(RequestBodyMinItemsIncreasedId, ERR, RequestPropertyMinItemsIncreasedCheck, DirectionRequest, LocationBody, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMinItemsIncreasedId, ERR, RequestPropertyMinItemsIncreasedCheck, DirectionRequest, LocationProperties, ActionIncrease),
		// RequestPropertyMinItemsSetCheck
		newBackwardCompatibilityRule(RequestBodyMinItemsSetId, WARN, RequestPropertyMinItemsSetCheck, DirectionRequest, LocationBody, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMinItemsSetId, WARN, RequestPropertyMinItemsSetCheck, DirectionRequest, LocationProperties, ActionSet),
		// RequestPropertyMinLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMinLengthIncreasedId, ERR, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, LocationBody, ActionIncrease),
		newBackwardCompatibilityRule(RequestBodyMinLengthDecreasedId, INFO, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, LocationBody, ActionDecrease),
		newBackwardCompatibilityRule(RequestPropertyMinLengthIncreasedId, ERR, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, LocationProperties, ActionIncrease),
		newBackwardCompatibilityRule(RequestPropertyMinLengthDecreasedId, INFO, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, LocationProperties, ActionDecrease),
		// RequestPropertyMinSetCheck
		newBackwardCompatibilityRule(RequestBodyMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, LocationBody, ActionSet),
		newBackwardCompatibilityRule(RequestPropertyMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, LocationProperties, ActionSet),
		// RequestPropertyOneOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyOneOfAddedId, INFO, RequestPropertyOneOfUpdatedCheck, DirectionRequest, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(RequestBodyOneOfRemovedId, ERR, RequestPropertyOneOfUpdatedCheck, DirectionRequest, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyOneOfAddedId, INFO, RequestPropertyOneOfUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyOneOfRemovedId, ERR, RequestPropertyOneOfUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		// RequestPropertyPatternUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyPatternRemovedId, INFO, RequestPropertyPatternUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(RequestPropertyPatternAddedId, WARN, RequestPropertyPatternUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(RequestPropertyPatternChangedId, WARN, RequestPropertyPatternUpdatedCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyPatternGeneralizedId, INFO, RequestPropertyPatternUpdatedCheck, DirectionRequest, LocationProperties, ActionGeneralize),
		// RequestPropertyRequiredUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyBecameRequiredId, ERR, RequestPropertyRequiredUpdatedCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecameRequiredWithDefaultId, INFO, RequestPropertyRequiredUpdatedCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyBecameOptionalId, INFO, RequestPropertyRequiredUpdatedCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestPropertyTypeChangedCheck
		newBackwardCompatibilityRule(RequestBodyTypeGeneralizedId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, LocationBody, ActionGeneralize),
		newBackwardCompatibilityRule(RequestBodyTypeChangedId, ERR, RequestPropertyTypeChangedCheck, DirectionRequest, LocationBody, ActionChange),
		newBackwardCompatibilityRule(RequestPropertyTypeGeneralizedId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, LocationProperties, ActionGeneralize),
		newBackwardCompatibilityRule(RequestPropertyTypeChangedId, ERR, RequestPropertyTypeChangedCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestPropertyUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyRemovedId, WARN, RequestPropertyUpdatedCheck, DirectionRequest, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(NewRequiredRequestPropertyId, ERR, RequestPropertyUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(NewRequiredRequestPropertyWithDefaultId, INFO, RequestPropertyUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(NewOptionalRequestPropertyId, INFO, RequestPropertyUpdatedCheck, DirectionRequest, LocationProperties, ActionAdd),
		// RequestPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameNonWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameNonReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameNonWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameNonReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, LocationProperties, ActionChange),
		// RequestPropertyXExtensibleEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestPropertyXExtensibleEnumValueRemovedId, ERR, RequestPropertyXExtensibleEnumValueRemovedCheck, DirectionRequest, LocationProperties, ActionRemove),
		// ResponseDiscriminatorUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorRemovedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorPropertyNameChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationBody, ActionChange),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingDeletedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationBody, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorRemovedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorPropertyNameChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingDeletedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, LocationProperties, ActionChange),
		// ResponseHeaderBecameOptionalCheck
		newBackwardCompatibilityRule(ResponseHeaderBecameOptionalId, ERR, ResponseHeaderBecameOptionalCheck, DirectionResponse, LocationHeaders, ActionChange),
		// ResponseHeaderRemovedCheck
		newBackwardCompatibilityRule(RequiredResponseHeaderRemovedId, ERR, ResponseHeaderRemovedCheck, DirectionResponse, LocationHeaders, ActionRemove),
		newBackwardCompatibilityRule(OptionalResponseHeaderRemovedId, WARN, ResponseHeaderRemovedCheck, DirectionResponse, LocationHeaders, ActionRemove),
		// ResponseMediaTypeUpdatedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeRemovedId, ERR, ResponseMediaTypeUpdatedCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponseMediaTypeAddedId, INFO, ResponseMediaTypeUpdatedCheck, DirectionResponse, LocationBody, ActionAdd),
		// ResponseMediaTypeNameUpdatedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeNameChangedId, INFO, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, LocationBody, ActionChange),
		newBackwardCompatibilityRule(ResponseMediaTypeNameGeneralizedId, ERR, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, LocationBody, ActionGeneralize),
		newBackwardCompatibilityRule(ResponseMediaTypeNameSpecializedId, INFO, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, LocationBody, ActionSpecialize),
		// ResponseOptionalPropertyUpdatedCheck
		newBackwardCompatibilityRule(ResponseOptionalPropertyRemovedId, WARN, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponseOptionalWriteOnlyPropertyRemovedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponseOptionalPropertyAddedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponseOptionalWriteOnlyPropertyAddedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		// ResponseOptionalPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameNonWriteOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameWriteOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameReadOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameNonReadOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		// ResponsePatternAddedOrChangedCheck
		newBackwardCompatibilityRule(ResponsePropertyPatternAddedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyPatternChangedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyPatternRemovedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, LocationProperties, ActionRemove),
		// ResponsePropertyAllOfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyAllOfAddedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyAllOfRemovedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyAllOfAddedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyAllOfRemovedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		// ResponsePropertyAnyOfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyAnyOfAddedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyAnyOfRemovedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyAnyOfAddedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyAnyOfRemovedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		// ResponsePropertyBecameNullableCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameNullableId, ERR, ResponsePropertyBecameNullableCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseBodyBecameNullableId, ERR, ResponsePropertyBecameNullableCheck, DirectionResponse, LocationBody, ActionChange),
		// ResponsePropertyBecameOptionalCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameOptionalId, ERR, ResponsePropertyBecameOptionalCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyBecameOptionalId, INFO, ResponsePropertyBecameOptionalCheck, DirectionResponse, LocationProperties, ActionChange),
		// ResponsePropertyBecameRequiredCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameRequiredId, INFO, ResponsePropertyBecameRequiredCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyBecameRequiredId, INFO, ResponsePropertyBecameRequiredCheck, DirectionResponse, LocationProperties, ActionChange),
		// ResponsePropertyDefaultValueChangedCheck
		newBackwardCompatibilityRule(ResponseBodyDefaultValueAddedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyDefaultValueRemovedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponseBodyDefaultValueChangedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, LocationBody, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueAddedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueRemovedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueChangedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, LocationProperties, ActionChange),
		// ResponsePropertyEnumValueAddedCheck
		newBackwardCompatibilityRule(ResponsePropertyEnumValueAddedId, WARN, ResponsePropertyEnumValueAddedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyEnumValueAddedId, INFO, ResponsePropertyEnumValueAddedCheck, DirectionResponse, LocationProperties, ActionAdd),
		// ResponsePropertyMaxIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, LocationBody, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, LocationProperties, ActionIncrease),
		// ResponsePropertyMaxLengthIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxLengthIncreasedId, ERR, ResponsePropertyMaxLengthIncreasedCheck, DirectionResponse, LocationBody, ActionIncrease),
		newBackwardCompatibilityRule(ResponsePropertyMaxLengthIncreasedId, ERR, ResponsePropertyMaxLengthIncreasedCheck, DirectionResponse, LocationProperties, ActionIncrease),
		// ResponsePropertyMaxLengthUnsetCheck
		newBackwardCompatibilityRule(ResponseBodyMaxLengthUnsetId, ERR, ResponsePropertyMaxLengthUnsetCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyMaxLengthUnsetId, ERR, ResponsePropertyMaxLengthUnsetCheck, DirectionResponse, LocationProperties, ActionRemove),
		// ResponsePropertyMinDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, LocationBody, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, LocationProperties, ActionDecrease),
		// ResponsePropertyMinItemsDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinItemsDecreasedId, ERR, ResponsePropertyMinItemsDecreasedCheck, DirectionResponse, LocationBody, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyMinItemsDecreasedId, ERR, ResponsePropertyMinItemsDecreasedCheck, DirectionResponse, LocationProperties, ActionDecrease),
		// ResponsePropertyMinItemsUnsetCheck
		newBackwardCompatibilityRule(ResponseBodyMinItemsUnsetId, ERR, ResponsePropertyMinItemsUnsetCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyMinItemsUnsetId, ERR, ResponsePropertyMinItemsUnsetCheck, DirectionResponse, LocationProperties, ActionRemove),
		// ResponsePropertyMinLengthDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinLengthDecreasedId, ERR, ResponsePropertyMinLengthDecreasedCheck, DirectionResponse, LocationBody, ActionDecrease),
		newBackwardCompatibilityRule(ResponsePropertyMinLengthDecreasedId, ERR, ResponsePropertyMinLengthDecreasedCheck, DirectionResponse, LocationProperties, ActionDecrease),
		// ResponsePropertyOneOfUpdated
		newBackwardCompatibilityRule(ResponseBodyOneOfAddedId, ERR, ResponsePropertyOneOfUpdated, DirectionResponse, LocationBody, ActionAdd),
		newBackwardCompatibilityRule(ResponseBodyOneOfRemovedId, INFO, ResponsePropertyOneOfUpdated, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyOneOfAddedId, ERR, ResponsePropertyOneOfUpdated, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponsePropertyOneOfRemovedId, INFO, ResponsePropertyOneOfUpdated, DirectionResponse, LocationProperties, ActionRemove),
		// ResponsePropertyTypeChangedCheck
		newBackwardCompatibilityRule(ResponseBodyTypeChangedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, LocationBody, ActionChange),
		newBackwardCompatibilityRule(ResponsePropertyTypeChangedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, LocationProperties, ActionChange),
		// ResponseRequiredPropertyUpdatedCheck
		newBackwardCompatibilityRule(ResponseRequiredPropertyRemovedId, ERR, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponseRequiredWriteOnlyPropertyRemovedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponseRequiredPropertyAddedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		newBackwardCompatibilityRule(ResponseRequiredWriteOnlyPropertyAddedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		// ResponseRequiredPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameNonWriteOnlyId, WARN, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameWriteOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameReadOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameNonReadOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, LocationProperties, ActionChange),
		// ResponseSuccessStatusUpdatedCheck
		newBackwardCompatibilityRule(ResponseSuccessStatusRemovedId, ERR, ResponseSuccessStatusUpdatedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponseSuccessStatusAddedId, INFO, ResponseSuccessStatusUpdatedCheck, DirectionResponse, LocationProperties, ActionAdd),
		// ResponseNonSuccessStatusUpdatedCheck
		newBackwardCompatibilityRule(ResponseNonSuccessStatusRemovedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, LocationNone, ActionRemove), // optional
		newBackwardCompatibilityRule(ResponseNonSuccessStatusAddedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, LocationNone, ActionAdd),
		// APIOperationIdUpdatedCheck
		newBackwardCompatibilityRule(APIOperationIdRemovedId, INFO, APIOperationIdUpdatedCheck, DirectionNone, LocationNone, ActionRemove), // optional
		newBackwardCompatibilityRule(APIOperationIdAddId, INFO, APIOperationIdUpdatedCheck, DirectionNone, LocationNone, ActionAdd),
		// APITagUpdatedCheck
		newBackwardCompatibilityRule(APITagRemovedId, INFO, APITagUpdatedCheck, DirectionNone, LocationNone, ActionRemove), // optional
		newBackwardCompatibilityRule(APITagAddedId, INFO, APITagUpdatedCheck, DirectionNone, LocationNone, ActionAdd),
		// APIComponentsSchemaRemovedCheck
		newBackwardCompatibilityRule(APISchemasRemovedId, INFO, APIComponentsSchemaRemovedCheck, DirectionNone, LocationComponents, ActionRemove), // optional
		// ResponseParameterEnumValueRemovedCheck
		newBackwardCompatibilityRule(ResponsePropertyEnumValueRemovedId, INFO, ResponseParameterEnumValueRemovedCheck, DirectionResponse, LocationProperties, ActionRemove), // optional
		// ResponseMediaTypeEnumValueRemovedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeEnumValueRemovedId, INFO, ResponseMediaTypeEnumValueRemovedCheck, DirectionResponse, LocationBody, ActionRemove), // optional
		// RequestBodyEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestBodyEnumValueRemovedId, INFO, RequestBodyEnumValueRemovedCheck, DirectionRequest, LocationBody, ActionRemove), // optional
	}
}

func GetOptionalRules() BackwardCompatibilityRules {
	return BackwardCompatibilityRules{
		newBackwardCompatibilityRule(ResponseNonSuccessStatusRemovedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APIOperationIdRemovedId, INFO, APIOperationIdUpdatedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APITagRemovedId, INFO, APITagUpdatedCheck, DirectionNone, LocationNone, ActionRemove),
		newBackwardCompatibilityRule(APISchemasRemovedId, INFO, APIComponentsSchemaRemovedCheck, DirectionNone, LocationComponents, ActionRemove),
		newBackwardCompatibilityRule(ResponsePropertyEnumValueRemovedId, INFO, ResponseParameterEnumValueRemovedCheck, DirectionResponse, LocationProperties, ActionRemove),
		newBackwardCompatibilityRule(ResponseMediaTypeEnumValueRemovedId, INFO, ResponseMediaTypeEnumValueRemovedCheck, DirectionResponse, LocationBody, ActionRemove),
		newBackwardCompatibilityRule(RequestBodyEnumValueRemovedId, INFO, RequestBodyEnumValueRemovedCheck, DirectionRequest, LocationBody, ActionRemove),
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
