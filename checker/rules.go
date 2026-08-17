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
	// AreaInfo's rules judge the versioning policy, not the wire contract:
	// info is annotation-only, so nothing in it can break a client, but
	// info.version is the declared shape of the change and can contradict it.
	AreaInfo
	// AreaServers has no rules yet: server changes are treated as deployment
	// metadata. It exists so the area taxonomy covers every top-level
	// document section a future rule could target.
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
	// Locations are the metaschema cells this rule covers, as
	// "location:action[,action...]" claims (see metaschema.ParseClaim).
	// Empty means the rule is not yet mapped; rule_coverage_test.go reports
	// unmapped rules and uncovered cells.
	Locations []string
}

func newBackwardCompatibilityRule(id string, level Level, handler BackwardCompatibilityCheck,
	direction Direction,
	area Area,
	kind Kind,
	action Action,
	locations ...string) BackwardCompatibilityRule {
	return BackwardCompatibilityRule{
		Id:          id,
		Level:       level,
		Description: descriptionId(id),
		Handler:     handler,
		Direction:   direction,
		Area:        area,
		Kind:        kind,
		Action:      action,
		Locations:   locations,
	}
}

type BackwardCompatibilityRules []BackwardCompatibilityRule

func GetAllRules() BackwardCompatibilityRules {
	return BackwardCompatibilityRules{
		// Request property deprecation checks
		newBackwardCompatibilityRule(RequestPropertyDeprecatedId, INFO, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.requestBody.content.*.schema.deprecated:set"),
		newBackwardCompatibilityRule(RequestPropertyDeprecatedWithSunsetId, INFO, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.requestBody.content.*.schema.deprecated:set"),
		newBackwardCompatibilityRule(RequestPropertyDeprecatedSunsetMissingId, ERR, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.requestBody.content.*.schema.deprecated:set"),
		newBackwardCompatibilityRule(RequestPropertyDeprecatedInvalidId, ERR, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.requestBody.content.*.schema.deprecated:set", "paths.*.*.requestBody.content.*.schema.x-*:add,change"),
		newBackwardCompatibilityRule(RequestPropertyReactivatedId, INFO, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.requestBody.content.*.schema.deprecated:unset"),
		newBackwardCompatibilityRule(RequestPropertySunsetDateTooSmallId, ERR, RequestPropertyDeprecationCheck, DirectionRequest, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.requestBody.content.*.schema.deprecated:set", "paths.*.*.requestBody.content.*.schema.x-*:add,change"),
		// Response property deprecation checks
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedId, INFO, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.responses.*.content.*.schema.deprecated:set"),
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedWithSunsetId, INFO, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.responses.*.content.*.schema.deprecated:set"),
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedSunsetMissingId, ERR, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.responses.*.content.*.schema.deprecated:set"),
		newBackwardCompatibilityRule(ResponsePropertyDeprecatedInvalidId, ERR, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.responses.*.content.*.schema.deprecated:set", "paths.*.*.responses.*.content.*.schema.x-*:add,change"),
		newBackwardCompatibilityRule(ResponsePropertyReactivatedId, INFO, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.responses.*.content.*.schema.deprecated:unset"),
		newBackwardCompatibilityRule(ResponsePropertySunsetDateTooSmallId, ERR, ResponsePropertyDeprecationCheck, DirectionResponse, AreaSchema, KindLifecycle, ActionChange, "paths.*.*.responses.*.content.*.schema.deprecated:set", "paths.*.*.responses.*.content.*.schema.x-*:add,change"),
		// APIAddedCheck
		newBackwardCompatibilityRule(EndpointAddedId, INFO, APIAddedCheck, DirectionNone, AreaPaths, KindExistence, ActionAdd, "paths.*:add", "paths.*.*:add"),
		// APIComponentsSecurityUpdatedCheck
		newBackwardCompatibilityRule(APIComponentsSecurityRemovedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove, "components.securitySchemes.*:remove"),
		newBackwardCompatibilityRule(APIComponentsSecurityAddedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd, "components.securitySchemes.*:add"),
		newBackwardCompatibilityRule(APIComponentsSecurityComponentOauthUrlUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange, "components.securitySchemes.*.flows.**.authorizationUrl:set,unset,change"),
		newBackwardCompatibilityRule(APIComponentsSecurityTypeUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange, "components.securitySchemes.*.type:change"),
		newBackwardCompatibilityRule(APIComponentsSecurityOauthTokenUrlUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange, "components.securitySchemes.*.flows.**.tokenUrl:set,unset,change"),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeAddedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd, "components.securitySchemes.*.flows.**.scopes.*:add"),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeRemovedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove, "components.securitySchemes.*.flows.**.scopes.*:remove"),
		newBackwardCompatibilityRule(APIComponentSecurityOauthScopeUpdatedId, INFO, APIComponentsSecurityUpdatedCheck, DirectionNone, AreaSecurity, KindType, ActionChange, "components.securitySchemes.*.flows.**.scopes.*:change"),
		// APISecurityUpdatedCheck
		newBackwardCompatibilityRule(APISecurityRemovedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove, "paths.*.*.security.*:remove"),
		newBackwardCompatibilityRule(APISecurityAddedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd, "paths.*.*.security.*:add"),
		newBackwardCompatibilityRule(APISecurityScopeAddedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd, "paths.*.*.security.*.*:add"),
		newBackwardCompatibilityRule(APISecurityScopeRemovedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove, "paths.*.*.security.*.*:remove"),
		newBackwardCompatibilityRule(APIGlobalSecurityRemovedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove, "security.*:remove"),
		newBackwardCompatibilityRule(APIGlobalSecurityAddedCheckId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd, "security.*:add"),
		newBackwardCompatibilityRule(APIGlobalSecurityScopeAddedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionAdd, "security.*.*:add"),
		newBackwardCompatibilityRule(APIGlobalSecurityScopeRemovedId, INFO, APISecurityUpdatedCheck, DirectionNone, AreaSecurity, KindExistence, ActionRemove, "security.*.*:remove"),
		// Versioning policy: run as part of CheckBackwardCompatibility, after
		// the checks, since it judges info.version against what they found.
		// INFO by default so it is quiet for teams that don't version with
		// semver; raise with --severity-levels to enforce.
		newBackwardCompatibilityRule(APIVersionNotBumpedId, INFO, nil, DirectionNone, AreaInfo, KindLifecycle, ActionChange, "info.version:change"),
		newBackwardCompatibilityRule(APIVersionDecreasedId, INFO, nil, DirectionNone, AreaInfo, KindLifecycle, ActionDecrease, "info.version:change"),
		newBackwardCompatibilityRule(APIMajorVersionNotBumpedId, INFO, nil, DirectionNone, AreaInfo, KindLifecycle, ActionIncrease, "info.version:change"),
		// Stability checks are run as part of CheckBackwardCompatibility.
		newBackwardCompatibilityRule(APIStabilityDecreasedId, ERR, nil, DirectionNone, AreaPaths, KindLifecycle, ActionDecrease, "paths.*.*.x-*:add,change"),
		newBackwardCompatibilityRule(APIStabilityIncreasedId, INFO, nil, DirectionNone, AreaPaths, KindLifecycle, ActionIncrease, "paths.*.*.x-*:add,change"),
		newBackwardCompatibilityRule(RequestPropertyStabilityDecreasedId, ERR, nil, DirectionRequest, AreaSchema, KindLifecycle, ActionDecrease, "paths.*.*.requestBody.content.*.schema.x-*:add,change"),
		newBackwardCompatibilityRule(RequestPropertyStabilityIncreasedId, INFO, nil, DirectionRequest, AreaSchema, KindLifecycle, ActionIncrease, "paths.*.*.requestBody.content.*.schema.x-*:add,change"),
		newBackwardCompatibilityRule(ResponsePropertyStabilityDecreasedId, ERR, nil, DirectionResponse, AreaSchema, KindLifecycle, ActionDecrease, "paths.*.*.responses.*.content.*.schema.x-*:add,change"),
		newBackwardCompatibilityRule(ResponsePropertyStabilityIncreasedId, INFO, nil, DirectionResponse, AreaSchema, KindLifecycle, ActionIncrease, "paths.*.*.responses.*.content.*.schema.x-*:add,change"),
		// APIDeprecationCheck
		newBackwardCompatibilityRule(EndpointReactivatedId, INFO, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.deprecated:unset"),
		newBackwardCompatibilityRule(APIDeprecatedSunsetParseId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.deprecated:set", "paths.*.*.x-*:add,change"),
		newBackwardCompatibilityRule(APIDeprecatedSunsetMissingId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.deprecated:set"),
		newBackwardCompatibilityRule(APIInvalidStabilityLevelId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.x-*:add,change"),
		newBackwardCompatibilityRule(APISunsetDateTooSmallId, ERR, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.deprecated:set", "paths.*.*.x-*:add,change"),
		newBackwardCompatibilityRule(EndpointDeprecatedId, INFO, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.deprecated:set"),
		newBackwardCompatibilityRule(EndpointDeprecatedWithSunsetId, INFO, APIDeprecationCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.deprecated:set"),
		// RequestParameterDeprecationCheck
		newBackwardCompatibilityRule(RequestParameterReactivatedId, INFO, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange, "paths.*.*.parameters.*.deprecated:unset"),
		newBackwardCompatibilityRule(RequestParameterDeprecatedSunsetMissingId, ERR, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange, "paths.*.*.parameters.*.deprecated:set"),
		newBackwardCompatibilityRule(RequestParameterSunsetDateTooSmallId, ERR, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange, "paths.*.*.parameters.*.deprecated:set", "paths.*.*.parameters.*.x-*:add,change"),
		newBackwardCompatibilityRule(RequestParameterDeprecatedId, INFO, RequestParameterDeprecationCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange, "paths.*.*.parameters.*.deprecated:set"),
		// APIRemovedCheck
		newBackwardCompatibilityRule(APIPathRemovedWithoutDeprecationId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove, "paths.*:remove"),
		newBackwardCompatibilityRule(APIPathRemovedWithDeprecationId, INFO, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove, "paths.*:remove"),
		newBackwardCompatibilityRule(APIPathSunsetParseId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*:remove"),
		newBackwardCompatibilityRule(APIPathRemovedBeforeSunsetId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove, "paths.*:remove"),
		newBackwardCompatibilityRule(APIRemovedWithoutDeprecationId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove, "paths.*.*:remove"),
		newBackwardCompatibilityRule(APIRemovedWithDeprecationId, INFO, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove, "paths.*.*:remove"),
		newBackwardCompatibilityRule(APIRemovedBeforeSunsetId, ERR, APIRemovedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove, "paths.*.*:remove"),
		// APISunsetChangedCheck
		newBackwardCompatibilityRule(APISunsetDeletedId, ERR, APISunsetChangedCheck, DirectionNone, AreaPaths, KindLifecycle, ActionRemove, "paths.*.*.x-*:remove"),
		newBackwardCompatibilityRule(APISunsetDateChangedTooSmallId, ERR, APISunsetChangedCheck, DirectionNone, AreaPaths, KindLifecycle, ActionChange, "paths.*.*.x-*:change"),
		// RequestParameterSunsetChangedCheck
		newBackwardCompatibilityRule(RequestParameterSunsetDeletedId, ERR, RequestParameterSunsetChangedCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange, "paths.*.*.parameters.*.x-*:remove"),
		newBackwardCompatibilityRule(RequestParameterSunsetDateChangedTooSmallId, ERR, RequestParameterSunsetChangedCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange, "paths.*.*.parameters.*.x-*:change"),
		// AddedRequiredRequestBodyCheck
		newBackwardCompatibilityRule(AddedRequiredRequestBodyId, ERR, AddedRequestBodyCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd, "paths.*.*.requestBody:set"),
		newBackwardCompatibilityRule(AddedOptionalRequestBodyId, INFO, AddedRequestBodyCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd, "paths.*.*.requestBody:set"),
		// NewRequestNonPathDefaultParameterCheck
		newBackwardCompatibilityRule(NewRequiredRequestDefaultParameterToExistingPathId, ERR, NewRequestNonPathDefaultParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd, "paths.*.parameters.*:add"),
		newBackwardCompatibilityRule(NewOptionalRequestDefaultParameterToExistingPathId, INFO, NewRequestNonPathDefaultParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd, "paths.*.parameters.*:add"),
		// NewRequestNonPathParameterCheck
		newBackwardCompatibilityRule(NewRequiredRequestParameterId, ERR, NewRequestNonPathParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd, "paths.*.*.parameters.*:add"),
		newBackwardCompatibilityRule(NewOptionalRequestParameterId, INFO, NewRequestNonPathParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd, "paths.*.*.parameters.*:add"),
		// NewRequestPathParameterCheck
		newBackwardCompatibilityRule(NewRequestPathParameterId, ERR, NewRequestPathParameterCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd, "paths.*.*.parameters.*:add"),
		// NewRequiredRequestHeaderPropertyCheck
		newBackwardCompatibilityRule(NewRequiredRequestHeaderPropertyId, ERR, NewRequiredRequestHeaderPropertyCheck, DirectionRequest, AreaParameters, KindExistence, ActionAdd, "paths.*.*.parameters.*.schema.properties.*:add"),
		// RequestBodyBecameEnumCheck
		newBackwardCompatibilityRule(RequestBodyBecameEnumId, ERR, RequestBodyBecameEnumCheck, DirectionRequest, AreaSchema, KindValues, ActionChange, "paths.*.*.requestBody.content.*.schema.enum:add"),
		// RequestBodyMediaTypeChangedCheck
		newBackwardCompatibilityRule(RequestBodyMediaTypeAddedId, INFO, RequestBodyMediaTypeChangedCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*:add"),
		newBackwardCompatibilityRule(RequestBodyMediaTypeRemovedId, ERR, RequestBodyMediaTypeChangedCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionRemove, "paths.*.*.requestBody.content.*:remove"),
		// MediaTypeSchemaExistenceCheck: a schema appearing/disappearing within an existing media type.
		newBackwardCompatibilityRule(RequestBodyMediaTypeSchemaAddedId, ERR, MediaTypeSchemaExistenceCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*.schema:set"),
		newBackwardCompatibilityRule(RequestBodyMediaTypeSchemaRemovedId, INFO, MediaTypeSchemaExistenceCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionRemove, "paths.*.*.requestBody.content.*.schema:unset"),
		newBackwardCompatibilityRule(ResponseBodyMediaTypeSchemaAddedId, INFO, MediaTypeSchemaExistenceCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.schema:set"),
		newBackwardCompatibilityRule(ResponseBodyMediaTypeSchemaRemovedId, ERR, MediaTypeSchemaExistenceCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.schema:unset"),
		// MediaTypeSchemaExistenceCheck: the same, for the OpenAPI 3.2 itemSchema.
		newBackwardCompatibilityRule(RequestBodyMediaTypeItemSchemaAddedId, ERR, MediaTypeSchemaExistenceCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*.itemSchema:set"),
		newBackwardCompatibilityRule(RequestBodyMediaTypeItemSchemaRemovedId, INFO, MediaTypeSchemaExistenceCheck, DirectionRequest, AreaRequestBody, KindExistence, ActionRemove, "paths.*.*.requestBody.content.*.itemSchema:unset"),
		newBackwardCompatibilityRule(ResponseBodyMediaTypeItemSchemaAddedId, INFO, MediaTypeSchemaExistenceCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.itemSchema:set"),
		newBackwardCompatibilityRule(ResponseBodyMediaTypeItemSchemaRemovedId, ERR, MediaTypeSchemaExistenceCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.itemSchema:unset"),
		newBackwardCompatibilityRule(ResponseBodyMediaTypeItemSchemaRemovedUntypedId, ERR, MediaTypeSchemaExistenceCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.itemSchema:unset"),
		// RequestBodyRemovedCheck
		newBackwardCompatibilityRule(RequestBodyRemovedId, ERR, RequestBodyRemovedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove, "paths.*.*.requestBody:unset"),
		// RequestBodyRequiredUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyBecameOptionalId, INFO, RequestBodyRequiredUpdatedCheck, DirectionRequest, AreaRequestBody, KindRequiredness, ActionChange, "paths.*.*.requestBody.required:unset"),
		newBackwardCompatibilityRule(RequestBodyBecameRequiredId, ERR, RequestBodyRequiredUpdatedCheck, DirectionRequest, AreaRequestBody, KindRequiredness, ActionChange, "paths.*.*.requestBody.required:set"),
		// RequestDiscriminatorUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyDiscriminatorAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.discriminator:set"),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorRemovedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.discriminator:unset"),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorPropertyNameChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange, "paths.*.*.requestBody.content.*.schema.discriminator.propertyName:change"),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.discriminator.mapping.*:add"),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingDeletedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.discriminator.mapping.*:remove"),
		newBackwardCompatibilityRule(RequestBodyDiscriminatorMappingChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange, "paths.*.*.requestBody.content.*.schema.discriminator.mapping.*:change"),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.discriminator:set"),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorRemovedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.discriminator:unset"),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorPropertyNameChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange, "paths.*.*.requestBody.content.*.schema.discriminator.propertyName:change"),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingAddedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.discriminator.mapping.*:add"),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingDeletedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.discriminator.mapping.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyDiscriminatorMappingChangedId, INFO, RequestDiscriminatorUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange, "paths.*.*.requestBody.content.*.schema.discriminator.mapping.*:change"),
		// RequestHeaderPropertyBecameEnumCheck
		newBackwardCompatibilityRule(RequestHeaderPropertyBecameEnumId, ERR, RequestHeaderPropertyBecameEnumCheck, DirectionRequest, AreaParameters, KindValues, ActionChange, "paths.*.*.parameters.*.schema.enum:add"),
		// RequestHeaderPropertyBecameRequiredCheck
		newBackwardCompatibilityRule(RequestHeaderPropertyBecameRequiredId, ERR, RequestHeaderPropertyBecameRequiredCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange, "paths.*.*.parameters.*.schema.required:add"),
		// RequestParameterBecameEnumCheck
		newBackwardCompatibilityRule(RequestParameterBecameEnumId, ERR, RequestParameterBecameEnumCheck, DirectionRequest, AreaParameters, KindValues, ActionChange, "paths.*.*.parameters.*.schema.enum:add"),
		newBackwardCompatibilityRule(RequestParameterBecameNullableId, INFO, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange, "paths.*.*.parameters.*.schema.nullable:set", "paths.*.*.parameters.*.schema.type:add"),
		newBackwardCompatibilityRule(RequestParameterBecameNotNullableId, ERR, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange, "paths.*.*.parameters.*.schema.nullable:unset", "paths.*.*.parameters.*.schema.type:remove"),
		newBackwardCompatibilityRule(RequestParameterPropertyBecameNullableId, INFO, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange, "paths.*.*.parameters.*.schema.nullable:set", "paths.*.*.parameters.*.schema.type:add"),
		newBackwardCompatibilityRule(RequestParameterPropertyBecameNotNullableId, ERR, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange, "paths.*.*.parameters.*.schema.nullable:unset", "paths.*.*.parameters.*.schema.type:remove"),
		// RequestParameterDefaultValueChangedCheck
		newBackwardCompatibilityRule(RequestParameterDefaultValueChangedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, AreaParameters, KindValues, ActionChange, "paths.*.*.parameters.*.schema.default:change"),
		newBackwardCompatibilityRule(RequestParameterDefaultValueAddedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, AreaParameters, KindValues, ActionAdd, "paths.*.*.parameters.*.schema.default:set"),
		newBackwardCompatibilityRule(RequestParameterDefaultValueRemovedId, ERR, RequestParameterDefaultValueChangedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove, "paths.*.*.parameters.*.schema.default:unset"),
		// RequestParameterEnumValueUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterEnumValueAddedId, INFO, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionAdd, "paths.*.*.parameters.*.schema.enum:add"),
		newBackwardCompatibilityRule(RequestParameterEnumValueRemovedId, ERR, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove, "paths.*.*.parameters.*.schema.enum:remove"),
		newBackwardCompatibilityRule(RequestParameterPropertyEnumValueAddedId, INFO, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionAdd, "paths.*.*.parameters.*.schema.enum:add"),
		newBackwardCompatibilityRule(RequestParameterPropertyEnumValueRemovedId, ERR, RequestParameterEnumValueUpdatedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove, "paths.*.*.parameters.*.schema.enum:remove"),
		// RequestParameterMaxItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxItemsIncreasedId, INFO, RequestParameterMaxItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.maxItems:increase"),
		newBackwardCompatibilityRule(RequestParameterMaxItemsDecreasedId, ERR, RequestParameterMaxItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.maxItems:decrease"),
		// RequestParameterMaxLengthSetCheck
		newBackwardCompatibilityRule(RequestParameterMaxLengthSetId, WARN, RequestParameterMaxLengthSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet, "paths.*.*.parameters.*.schema.maxLength:set"),
		// RequestParameterMaxLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxLengthIncreasedId, INFO, RequestParameterMaxLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.maxLength:increase"),
		newBackwardCompatibilityRule(RequestParameterMaxLengthDecreasedId, ERR, RequestParameterMaxLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.maxLength:decrease"),
		// RequestParameterMaxSetCheck
		newBackwardCompatibilityRule(RequestParameterMaxSetId, WARN, RequestParameterMaxSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet, "paths.*.*.parameters.*.schema.maximum:set"),
		newBackwardCompatibilityRule(RequestParameterExclusiveMaxSetId, WARN, RequestParameterMaxSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet, "paths.*.*.parameters.*.schema.exclusiveMaximum:set"),
		// RequestParameterMaxUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMaxIncreasedId, INFO, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.maximum:increase"),
		newBackwardCompatibilityRule(RequestParameterMaxDecreasedId, ERR, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.maximum:decrease"),
		newBackwardCompatibilityRule(RequestParameterExclusiveMaxIncreasedId, INFO, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.exclusiveMaximum:increase"),
		newBackwardCompatibilityRule(RequestParameterExclusiveMaxDecreasedId, ERR, RequestParameterMaxUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.exclusiveMaximum:decrease"),
		// RequestParameterMinItemsSetCheck
		newBackwardCompatibilityRule(RequestParameterMinItemsSetId, WARN, RequestParameterMinItemsSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet, "paths.*.*.parameters.*.schema.minItems:set"),
		// RequestParameterMinItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinItemsIncreasedId, ERR, RequestParameterMinItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.minItems:increase"),
		newBackwardCompatibilityRule(RequestParameterMinItemsDecreasedId, INFO, RequestParameterMinItemsUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.minItems:decrease"),
		// RequestParameterMinLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinLengthIncreasedId, ERR, RequestParameterMinLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.minLength:increase"),
		newBackwardCompatibilityRule(RequestParameterMinLengthDecreasedId, INFO, RequestParameterMinLengthUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.minLength:decrease"),
		// RequestParameterMinSetCheck
		newBackwardCompatibilityRule(RequestParameterMinSetId, WARN, RequestParameterMinSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet, "paths.*.*.parameters.*.schema.minimum:set"),
		newBackwardCompatibilityRule(RequestParameterExclusiveMinSetId, WARN, RequestParameterMinSetCheck, DirectionRequest, AreaParameters, KindConstraints, ActionSet, "paths.*.*.parameters.*.schema.exclusiveMinimum:set"),
		// RequestParameterMinUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterMinIncreasedId, ERR, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.minimum:increase"),
		newBackwardCompatibilityRule(RequestParameterMinDecreasedId, INFO, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.minimum:decrease"),
		newBackwardCompatibilityRule(RequestParameterExclusiveMinIncreasedId, ERR, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionIncrease, "paths.*.*.parameters.*.schema.exclusiveMinimum:increase"),
		newBackwardCompatibilityRule(RequestParameterExclusiveMinDecreasedId, INFO, RequestParameterMinUpdatedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionDecrease, "paths.*.*.parameters.*.schema.exclusiveMinimum:decrease"),
		// RequestParameterPatternAddedOrChangedCheck
		newBackwardCompatibilityRule(RequestParameterPatternAddedId, ERR, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionAdd, "paths.*.*.parameters.*.schema.pattern:set"),
		newBackwardCompatibilityRule(RequestParameterPatternRemovedId, INFO, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionRemove, "paths.*.*.parameters.*.schema.pattern:unset"),
		newBackwardCompatibilityRule(RequestParameterPatternChangedId, WARN, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionChange, "paths.*.*.parameters.*.schema.pattern:change"),
		newBackwardCompatibilityRule(RequestParameterPatternGeneralizedId, INFO, RequestParameterPatternAddedOrChangedCheck, DirectionRequest, AreaParameters, KindConstraints, ActionGeneralize, "paths.*.*.parameters.*.schema.pattern:change"),
		// RequestParameterRemovedCheck
		newBackwardCompatibilityRule(RequestParameterRemovedId, WARN, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindExistence, ActionRemove, "paths.*.*.parameters.*:remove"),
		newBackwardCompatibilityRule(RequestParameterRemovedWithDeprecationId, INFO, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindExistence, ActionRemove, "paths.*.*.parameters.*:remove"),
		newBackwardCompatibilityRule(RequestParameterSunsetParseId, ERR, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindLifecycle, ActionChange, "paths.*.*.parameters.*:remove"),
		newBackwardCompatibilityRule(ParameterRemovedBeforeSunsetId, ERR, RequestParameterRemovedCheck, DirectionRequest, AreaParameters, KindExistence, ActionRemove, "paths.*.*.parameters.*:remove"),
		// RequestParameterRequiredValueUpdatedCheck
		newBackwardCompatibilityRule(RequestParameterBecomeRequiredId, ERR, RequestParameterRequiredValueUpdatedCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange, "paths.*.*.parameters.*.required:set"),
		newBackwardCompatibilityRule(RequestParameterBecomeOptionalId, INFO, RequestParameterRequiredValueUpdatedCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange, "paths.*.*.parameters.*.required:unset"),
		// RequestParameterTypeChangedCheck
		newBackwardCompatibilityRule(RequestParameterTypeChangedId, ERR, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionChange, "paths.*.*.parameters.*.schema.type:add,remove", "paths.*.*.parameters.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(RequestParameterTypeGeneralizedId, INFO, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionGeneralize, "paths.*.*.parameters.*.schema.type:add", "paths.*.*.parameters.*.schema.format:unset,change"),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeChangedId, WARN, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionChange, "paths.*.*.parameters.*.schema.type:add,remove", "paths.*.*.parameters.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeGeneralizedId, INFO, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionGeneralize, "paths.*.*.parameters.*.schema.type:add", "paths.*.*.parameters.*.schema.format:unset,change"),
		newBackwardCompatibilityRule(RequestParameterPropertyTypeSpecializedId, ERR, RequestParameterTypeChangedCheck, DirectionRequest, AreaParameters, KindType, ActionSpecialize, "paths.*.*.parameters.*.schema.type:remove", "paths.*.*.parameters.*.schema.format:set,change"),
		// RequestParameterXExtensibleEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestParameterXExtensibleEnumValueRemovedId, ERR, RequestParameterXExtensibleEnumValueRemovedCheck, DirectionRequest, AreaParameters, KindValues, ActionRemove, "paths.*.*.parameters.*.schema.x-*:change"),
		// RequestPropertyAllOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyAllOfAddedId, ERR, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(RequestBodyAllOfRemovedId, WARN, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.allOf.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyAllOfAddedId, ERR, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(RequestPropertyAllOfRemovedId, WARN, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.allOf.*:remove"),
		newBackwardCompatibilityRule(RequestBodyAllOfAddedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(RequestBodyAllOfRemovedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.allOf.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyAllOfAddedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(RequestPropertyAllOfRemovedAnnotationOnlyId, INFO, RequestPropertyAllOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.allOf.*:remove"),
		// RequestPropertyAnyOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyAnyOfAddedId, INFO, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.anyOf.*:add"),
		newBackwardCompatibilityRule(RequestBodyAnyOfRemovedId, ERR, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.anyOf.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyAnyOfAddedId, INFO, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.anyOf.*:add"),
		newBackwardCompatibilityRule(RequestPropertyAnyOfRemovedId, ERR, RequestPropertyAnyOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.anyOf.*:remove"),
		// RequestPropertyBecameEnumCheck
		newBackwardCompatibilityRule(RequestPropertyBecameEnumId, ERR, RequestPropertyBecameEnumCheck, DirectionRequest, AreaSchema, KindValues, ActionChange, "paths.*.*.requestBody.content.*.schema.enum:add"),
		// RequestPropertyBecameNotNullableCheck
		newBackwardCompatibilityRule(RequestBodyBecomeNotNullableId, ERR, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.requestBody.content.*.schema.nullable:unset", "paths.*.*.requestBody.content.*.schema.type:remove"),
		newBackwardCompatibilityRule(RequestBodyBecomeNullableId, INFO, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.requestBody.content.*.schema.nullable:set", "paths.*.*.requestBody.content.*.schema.type:add"),
		newBackwardCompatibilityRule(RequestPropertyBecomeNotNullableId, ERR, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.requestBody.content.*.schema.nullable:unset", "paths.*.*.requestBody.content.*.schema.type:remove"),
		newBackwardCompatibilityRule(RequestPropertyBecomeNullableId, INFO, RequestPropertyBecameNotNullableCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.requestBody.content.*.schema.nullable:set", "paths.*.*.requestBody.content.*.schema.type:add"),
		// RequestPropertyDefaultValueChangedCheck
		newBackwardCompatibilityRule(RequestBodyDefaultValueAddedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd, "paths.*.*.requestBody.content.*.schema.default:set"),
		newBackwardCompatibilityRule(RequestBodyDefaultValueRemovedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.default:unset"),
		newBackwardCompatibilityRule(RequestBodyDefaultValueChangedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange, "paths.*.*.requestBody.content.*.schema.default:change"),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueAddedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd, "paths.*.*.requestBody.content.*.schema.default:set"),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueRemovedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.default:unset"),
		newBackwardCompatibilityRule(RequestPropertyDefaultValueChangedId, INFO, RequestPropertyDefaultValueChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange, "paths.*.*.requestBody.content.*.schema.default:change"),
		// RequestPropertyConstChangedCheck
		newBackwardCompatibilityRule(RequestBodyConstAddedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd, "paths.*.*.requestBody.content.*.schema.const:set"),
		newBackwardCompatibilityRule(RequestBodyConstRemovedId, INFO, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.const:unset"),
		newBackwardCompatibilityRule(RequestBodyConstChangedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange, "paths.*.*.requestBody.content.*.schema.const:change"),
		newBackwardCompatibilityRule(RequestPropertyConstAddedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd, "paths.*.*.requestBody.content.*.schema.const:set"),
		newBackwardCompatibilityRule(RequestPropertyConstRemovedId, INFO, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.const:unset"),
		newBackwardCompatibilityRule(RequestPropertyConstChangedId, ERR, RequestPropertyConstChangedCheck, DirectionRequest, AreaSchema, KindValues, ActionChange, "paths.*.*.requestBody.content.*.schema.const:change"),
		// RequestPropertyEnumValueUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyEnumValueRemovedId, ERR, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.enum:remove"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyEnumValueRemovedId, INFO, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.enum:remove"),
		newBackwardCompatibilityRule(RequestPropertyEnumValueAddedId, INFO, RequestPropertyEnumValueUpdatedCheck, DirectionRequest, AreaSchema, KindValues, ActionAdd, "paths.*.*.requestBody.content.*.schema.enum:add"),
		// RequestPropertyMaxDecreasedCheck
		newBackwardCompatibilityRule(RequestBodyMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maximum:decrease"),
		newBackwardCompatibilityRule(RequestBodyMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maximum:increase"),
		newBackwardCompatibilityRule(RequestPropertyMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maximum:decrease"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxDecreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maximum:decrease"),
		newBackwardCompatibilityRule(RequestPropertyMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maximum:increase"),
		newBackwardCompatibilityRule(RequestBodyExclusiveMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.exclusiveMaximum:decrease"),
		newBackwardCompatibilityRule(RequestBodyExclusiveMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.exclusiveMaximum:increase"),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMaxDecreasedId, ERR, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.exclusiveMaximum:decrease"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyExclusiveMaxDecreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.exclusiveMaximum:decrease"),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMaxIncreasedId, INFO, RequestPropertyMaxDecreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.exclusiveMaximum:increase"),
		// RequestPropertyMaxLengthSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxLengthSetId, WARN, RequestPropertyMaxLengthSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maxLength:set"),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthSetId, WARN, RequestPropertyMaxLengthSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maxLength:set"),
		// RequestPropertyMaxLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMaxLengthDecreasedId, ERR, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxLength:decrease"),
		newBackwardCompatibilityRule(RequestBodyMaxLengthIncreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxLength:increase"),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthDecreasedId, ERR, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxLength:decrease"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxLengthDecreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxLength:decrease"),
		newBackwardCompatibilityRule(RequestPropertyMaxLengthIncreasedId, INFO, RequestPropertyMaxLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxLength:increase"),
		// RequestPropertyMaxSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maximum:set"),
		newBackwardCompatibilityRule(RequestPropertyMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maximum:set"),
		newBackwardCompatibilityRule(RequestBodyExclusiveMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.exclusiveMaximum:set"),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMaxSetId, WARN, RequestPropertyMaxSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.exclusiveMaximum:set"),
		// RequestPropertyMinIncreasedCheck
		newBackwardCompatibilityRule(RequestBodyMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minimum:increase"),
		newBackwardCompatibilityRule(RequestBodyMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.minimum:decrease"),
		newBackwardCompatibilityRule(RequestPropertyMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minimum:increase"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMinIncreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minimum:increase"),
		newBackwardCompatibilityRule(RequestPropertyMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.minimum:decrease"),
		newBackwardCompatibilityRule(RequestBodyExclusiveMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.exclusiveMinimum:increase"),
		newBackwardCompatibilityRule(RequestBodyExclusiveMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.exclusiveMinimum:decrease"),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMinIncreasedId, ERR, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.exclusiveMinimum:increase"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyExclusiveMinIncreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.exclusiveMinimum:increase"),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMinDecreasedId, INFO, RequestPropertyMinIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.exclusiveMinimum:decrease"),
		// RequestPropertyMinItemsIncreasedCheck
		newBackwardCompatibilityRule(RequestBodyMinItemsIncreasedId, ERR, RequestPropertyMinItemsIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minItems:increase"),
		newBackwardCompatibilityRule(RequestPropertyMinItemsIncreasedId, ERR, RequestPropertyMinItemsIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minItems:increase"),
		// RequestPropertyMinItemsSetCheck
		newBackwardCompatibilityRule(RequestBodyMinItemsSetId, WARN, RequestPropertyMinItemsSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.minItems:set"),
		newBackwardCompatibilityRule(RequestPropertyMinItemsSetId, WARN, RequestPropertyMinItemsSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.minItems:set"),
		// RequestPropertyMinLengthUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMinLengthIncreasedId, ERR, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minLength:increase"),
		newBackwardCompatibilityRule(RequestBodyMinLengthDecreasedId, INFO, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.minLength:decrease"),
		newBackwardCompatibilityRule(RequestPropertyMinLengthIncreasedId, ERR, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minLength:increase"),
		newBackwardCompatibilityRule(RequestPropertyMinLengthDecreasedId, INFO, RequestPropertyMinLengthUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.minLength:decrease"),
		// RequestPropertyMinSetCheck
		newBackwardCompatibilityRule(RequestBodyMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.minimum:set"),
		newBackwardCompatibilityRule(RequestPropertyMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.minimum:set"),
		newBackwardCompatibilityRule(RequestBodyExclusiveMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.exclusiveMinimum:set"),
		newBackwardCompatibilityRule(RequestPropertyExclusiveMinSetId, WARN, RequestPropertyMinSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.exclusiveMinimum:set"),
		// RequestPropertyOneOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyOneOfAddedId, INFO, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.oneOf.*:add"),
		newBackwardCompatibilityRule(RequestBodyOneOfRemovedId, ERR, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.oneOf.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyOneOfAddedId, INFO, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.oneOf.*:add"),
		newBackwardCompatibilityRule(RequestPropertyOneOfRemovedId, ERR, RequestPropertyOneOfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.oneOf.*:remove"),
		// RequestPropertyPatternUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyPatternRemovedId, INFO, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.requestBody.content.*.schema.pattern:unset"),
		newBackwardCompatibilityRule(RequestPropertyPatternAddedId, ERR, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionAdd, "paths.*.*.requestBody.content.*.schema.pattern:set"),
		newBackwardCompatibilityRule(RequestPropertyPatternChangedId, WARN, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionChange, "paths.*.*.requestBody.content.*.schema.pattern:change"),
		newBackwardCompatibilityRule(RequestPropertyPatternGeneralizedId, INFO, RequestPropertyPatternUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionGeneralize, "paths.*.*.requestBody.content.*.schema.pattern:change"),
		// RequestPropertyRequiredUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyBecameRequiredId, ERR, RequestPropertyRequiredUpdatedCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.requestBody.content.*.schema.required:add"),
		newBackwardCompatibilityRule(RequestPropertyBecameRequiredWithDefaultId, ERR, RequestPropertyRequiredUpdatedCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.requestBody.content.*.schema.required:add"),
		newBackwardCompatibilityRule(RequestPropertyBecameOptionalId, INFO, RequestPropertyRequiredUpdatedCheck, DirectionRequest, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.requestBody.content.*.schema.required:remove"),
		// RequestPropertyTypeChangedCheck
		newBackwardCompatibilityRule(RequestBodyTypeGeneralizedId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionGeneralize, "paths.*.*.requestBody.content.*.schema.type:add", "paths.*.*.requestBody.content.*.schema.format:unset,change"),
		newBackwardCompatibilityRule(RequestBodyTypeCompatibleId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.type:add,remove", "paths.*.*.requestBody.content.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(RequestBodyTypeChangedId, ERR, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.type:add,remove", "paths.*.*.requestBody.content.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(RequestPropertyTypeGeneralizedId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionGeneralize, "paths.*.*.requestBody.content.*.schema.type:add", "paths.*.*.requestBody.content.*.schema.format:unset,change"),
		newBackwardCompatibilityRule(RequestPropertyTypeCompatibleId, INFO, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.type:add,remove", "paths.*.*.requestBody.content.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(RequestPropertyTypeChangedId, ERR, RequestPropertyTypeChangedCheck, DirectionRequest, AreaSchema, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.type:add,remove", "paths.*.*.requestBody.content.*.schema.format:set,unset,change"),
		// RequestPropertyUpdatedCheck
		newBackwardCompatibilityRule(RequestPropertyRemovedId, WARN, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove, "paths.*.*.requestBody.content.*.schema.properties.*:remove"),
		newBackwardCompatibilityRule(NewRequiredRequestPropertyId, ERR, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*.schema.properties.*:add"),
		newBackwardCompatibilityRule(NewRequiredRequestPropertyWithDefaultId, ERR, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*.schema.properties.*:add"),
		newBackwardCompatibilityRule(NewOptionalRequestPropertyId, INFO, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*.schema.properties.*:add"),
		newBackwardCompatibilityRule(RequestBodyWrappedInOneOfId, ERR, RequestPropertyUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange, "paths.*.*.requestBody.content.*.schema.oneOf.*:add"),
		// RequestPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameNonWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.writeOnly:unset"),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.writeOnly:set"),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.readOnly:set"),
		newBackwardCompatibilityRule(RequestOptionalPropertyBecameNonReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.readOnly:unset"),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameNonWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.writeOnly:unset"),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameWriteOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.writeOnly:set"),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.readOnly:set"),
		newBackwardCompatibilityRule(RequestRequiredPropertyBecameNonReadOnlyCheckId, INFO, RequestPropertyWriteOnlyReadOnlyCheck, DirectionRequest, AreaSchema, KindMutability, ActionChange, "paths.*.*.requestBody.content.*.schema.readOnly:unset"),
		// RequestPropertyXExtensibleEnumValueRemovedCheck
		newBackwardCompatibilityRule(RequestPropertyXExtensibleEnumValueRemovedId, ERR, RequestPropertyXExtensibleEnumValueRemovedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.x-*:change"),
		// ResponseDiscriminatorUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.discriminator:set"),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorRemovedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.discriminator:unset"),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorPropertyNameChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange, "paths.*.*.responses.*.content.*.schema.discriminator.propertyName:change"),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.discriminator.mapping.*:add"),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingDeletedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.discriminator.mapping.*:remove"),
		newBackwardCompatibilityRule(ResponseBodyDiscriminatorMappingChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange, "paths.*.*.responses.*.content.*.schema.discriminator.mapping.*:change"),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.discriminator:set"),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorRemovedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.discriminator:unset"),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorPropertyNameChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange, "paths.*.*.responses.*.content.*.schema.discriminator.propertyName:change"),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingAddedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.discriminator.mapping.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingDeletedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.discriminator.mapping.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyDiscriminatorMappingChangedId, INFO, ResponseDiscriminatorUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange, "paths.*.*.responses.*.content.*.schema.discriminator.mapping.*:change"),
		// ResponseHeaderBecameOptionalCheck
		newBackwardCompatibilityRule(ResponseHeaderBecameOptionalId, ERR, ResponseHeaderBecameOptionalCheck, DirectionResponse, AreaHeaders, KindRequiredness, ActionChange, "paths.*.*.responses.*.headers.*.required:unset"),
		// ResponseHeaderRemovedCheck
		newBackwardCompatibilityRule(RequiredResponseHeaderRemovedId, ERR, ResponseHeaderRemovedCheck, DirectionResponse, AreaHeaders, KindExistence, ActionRemove, "paths.*.*.responses.*.headers.*:remove"),
		newBackwardCompatibilityRule(OptionalResponseHeaderRemovedId, WARN, ResponseHeaderRemovedCheck, DirectionResponse, AreaHeaders, KindExistence, ActionRemove, "paths.*.*.responses.*.headers.*:remove"),
		// ResponseHeaderAddedCheck
		newBackwardCompatibilityRule(ResponseHeaderAddedId, INFO, ResponseHeaderAddedCheck, DirectionResponse, AreaHeaders, KindExistence, ActionAdd, "paths.*.*.responses.*.headers.*:add"),
		// ResponseHeaderTypeChangedCheck
		newBackwardCompatibilityRule(ResponseHeaderTypeChangedId, ERR, ResponseHeaderTypeChangedCheck, DirectionResponse, AreaHeaders, KindType, ActionChange, "paths.*.*.responses.*.headers.*.schema.type:add,remove", "paths.*.*.responses.*.headers.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(ResponseHeaderTypeGeneralizedId, ERR, ResponseHeaderTypeChangedCheck, DirectionResponse, AreaHeaders, KindType, ActionGeneralize, "paths.*.*.responses.*.headers.*.schema.type:add", "paths.*.*.responses.*.headers.*.schema.format:unset,change"),
		newBackwardCompatibilityRule(ResponseHeaderTypeSpecializedId, INFO, ResponseHeaderTypeChangedCheck, DirectionResponse, AreaHeaders, KindType, ActionSpecialize, "paths.*.*.responses.*.headers.*.schema.type:remove", "paths.*.*.responses.*.headers.*.schema.format:set,change"),
		newBackwardCompatibilityRule(ResponseHeaderTypeCompatibleId, INFO, ResponseHeaderTypeChangedCheck, DirectionResponse, AreaHeaders, KindType, ActionChange, "paths.*.*.responses.*.headers.*.schema.type:add,remove", "paths.*.*.responses.*.headers.*.schema.format:set,unset,change"),
		// ResponseHeaderBecameNullableCheck
		newBackwardCompatibilityRule(ResponseHeaderBecameNullableId, ERR, ResponseHeaderBecameNullableCheck, DirectionResponse, AreaHeaders, KindRequiredness, ActionChange, "paths.*.*.responses.*.headers.*.schema.nullable:set", "paths.*.*.responses.*.headers.*.schema.type:add"),
		newBackwardCompatibilityRule(ResponseHeaderBecameNotNullableId, INFO, ResponseHeaderBecameNullableCheck, DirectionResponse, AreaHeaders, KindRequiredness, ActionChange, "paths.*.*.responses.*.headers.*.schema.nullable:unset", "paths.*.*.responses.*.headers.*.schema.type:remove"),
		// ResponseMediaTypeUpdatedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeRemovedId, ERR, ResponseMediaTypeUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*:remove"),
		newBackwardCompatibilityRule(ResponseMediaTypeAddedId, INFO, ResponseMediaTypeUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*:add"),
		// ResponseMediaTypeNameUpdatedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeNameChangedId, INFO, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionChange, "paths.*.*.responses.*.content.*:add,remove"),
		newBackwardCompatibilityRule(ResponseMediaTypeNameGeneralizedId, ERR, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionGeneralize, "paths.*.*.responses.*.content.*:add,remove"),
		newBackwardCompatibilityRule(ResponseMediaTypeNameSpecializedId, INFO, ResponseMediaTypeNameUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionSpecialize, "paths.*.*.responses.*.content.*:add,remove"),
		// ResponseOptionalPropertyUpdatedCheck
		newBackwardCompatibilityRule(ResponseOptionalPropertyRemovedId, WARN, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.schema.properties.*:remove"),
		newBackwardCompatibilityRule(ResponseOptionalWriteOnlyPropertyRemovedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.schema.properties.*:remove"),
		newBackwardCompatibilityRule(ResponseOptionalPropertyAddedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.schema.properties.*:add"),
		newBackwardCompatibilityRule(ResponseOptionalWriteOnlyPropertyAddedId, INFO, ResponseOptionalPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.schema.properties.*:add"),
		// ResponseOptionalPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameNonWriteOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.writeOnly:unset"),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameWriteOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.writeOnly:set"),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameReadOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.readOnly:set"),
		newBackwardCompatibilityRule(ResponseOptionalPropertyBecameNonReadOnlyId, INFO, ResponseOptionalPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.readOnly:unset"),
		// ResponsePatternAddedOrChangedCheck
		newBackwardCompatibilityRule(ResponsePropertyPatternAddedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionAdd, "paths.*.*.responses.*.content.*.schema.pattern:set"),
		newBackwardCompatibilityRule(ResponsePropertyPatternChangedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionChange, "paths.*.*.responses.*.content.*.schema.pattern:change"),
		newBackwardCompatibilityRule(ResponsePropertyPatternRemovedId, INFO, ResponsePatternAddedOrChangedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.pattern:unset"),
		// ResponsePropertyAllOfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyAllOfAddedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(ResponseBodyAllOfRemovedId, ERR, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.allOf.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyAllOfAddedId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyAllOfRemovedId, ERR, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.allOf.*:remove"),
		newBackwardCompatibilityRule(ResponseBodyAllOfAddedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(ResponseBodyAllOfRemovedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.allOf.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyAllOfAddedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.allOf.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyAllOfRemovedAnnotationOnlyId, INFO, ResponsePropertyAllOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.allOf.*:remove"),
		// ResponsePropertyAnyOfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyAnyOfAddedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.anyOf.*:add"),
		newBackwardCompatibilityRule(ResponseBodyAnyOfRemovedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.anyOf.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyAnyOfAddedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.anyOf.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyAnyOfRemovedId, INFO, ResponsePropertyAnyOfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.anyOf.*:remove"),
		// ResponsePropertyBecameNullableCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameNullableId, ERR, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.nullable:set", "paths.*.*.responses.*.content.*.schema.type:add"),
		newBackwardCompatibilityRule(ResponseBodyBecameNullableId, ERR, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.nullable:set", "paths.*.*.responses.*.content.*.schema.type:add"),
		newBackwardCompatibilityRule(ResponsePropertyBecameNotNullableId, INFO, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.nullable:unset", "paths.*.*.responses.*.content.*.schema.type:remove"),
		newBackwardCompatibilityRule(ResponseBodyBecameNotNullableId, INFO, ResponsePropertyBecameNullableCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.nullable:unset", "paths.*.*.responses.*.content.*.schema.type:remove"),
		// ResponsePropertyBecameOptionalCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameOptionalId, ERR, ResponsePropertyBecameOptionalCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.required:remove"),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyBecameOptionalId, INFO, ResponsePropertyBecameOptionalCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.required:remove"),
		// ResponsePropertyBecameRequiredCheck
		newBackwardCompatibilityRule(ResponsePropertyBecameRequiredId, INFO, ResponsePropertyBecameRequiredCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.required:add"),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyBecameRequiredId, INFO, ResponsePropertyBecameRequiredCheck, DirectionResponse, AreaSchema, KindRequiredness, ActionChange, "paths.*.*.responses.*.content.*.schema.required:add"),
		// ResponsePropertyDefaultValueChangedCheck
		newBackwardCompatibilityRule(ResponseBodyDefaultValueAddedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd, "paths.*.*.responses.*.content.*.schema.default:set"),
		newBackwardCompatibilityRule(ResponseBodyDefaultValueRemovedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove, "paths.*.*.responses.*.content.*.schema.default:unset"),
		newBackwardCompatibilityRule(ResponseBodyDefaultValueChangedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange, "paths.*.*.responses.*.content.*.schema.default:change"),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueAddedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd, "paths.*.*.responses.*.content.*.schema.default:set"),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueRemovedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove, "paths.*.*.responses.*.content.*.schema.default:unset"),
		newBackwardCompatibilityRule(ResponsePropertyDefaultValueChangedId, INFO, ResponsePropertyDefaultValueChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange, "paths.*.*.responses.*.content.*.schema.default:change"),
		// ResponsePropertyConstChangedCheck
		newBackwardCompatibilityRule(ResponseBodyConstAddedId, INFO, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd, "paths.*.*.responses.*.content.*.schema.const:set"),
		newBackwardCompatibilityRule(ResponseBodyConstRemovedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove, "paths.*.*.responses.*.content.*.schema.const:unset"),
		newBackwardCompatibilityRule(ResponseBodyConstChangedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange, "paths.*.*.responses.*.content.*.schema.const:change"),
		newBackwardCompatibilityRule(ResponsePropertyConstAddedId, INFO, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd, "paths.*.*.responses.*.content.*.schema.const:set"),
		newBackwardCompatibilityRule(ResponsePropertyConstRemovedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove, "paths.*.*.responses.*.content.*.schema.const:unset"),
		newBackwardCompatibilityRule(ResponsePropertyConstChangedId, ERR, ResponsePropertyConstChangedCheck, DirectionResponse, AreaSchema, KindValues, ActionChange, "paths.*.*.responses.*.content.*.schema.const:change"),
		// ResponsePropertyEnumValueAddedCheck
		newBackwardCompatibilityRule(ResponsePropertyEnumValueAddedId, WARN, ResponsePropertyEnumValueAddedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd, "paths.*.*.responses.*.content.*.schema.enum:add"),
		newBackwardCompatibilityRule(ResponseWriteOnlyPropertyEnumValueAddedId, INFO, ResponsePropertyEnumValueAddedCheck, DirectionResponse, AreaSchema, KindValues, ActionAdd, "paths.*.*.responses.*.content.*.schema.enum:add"),
		// ResponsePropertyMaxIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maximum:increase"),
		newBackwardCompatibilityRule(ResponsePropertyMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maximum:increase"),
		newBackwardCompatibilityRule(ResponseBodyExclusiveMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.exclusiveMaximum:increase"),
		newBackwardCompatibilityRule(ResponsePropertyExclusiveMaxIncreasedId, ERR, ResponsePropertyMaxIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.exclusiveMaximum:increase"),
		// ResponsePropertyMaxLengthIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxLengthIncreasedId, ERR, ResponsePropertyMaxLengthIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxLength:increase"),
		newBackwardCompatibilityRule(ResponsePropertyMaxLengthIncreasedId, ERR, ResponsePropertyMaxLengthIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxLength:increase"),
		// ResponsePropertyMaxLengthUnsetCheck
		newBackwardCompatibilityRule(ResponseBodyMaxLengthUnsetId, ERR, ResponsePropertyMaxLengthUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.maxLength:unset"),
		newBackwardCompatibilityRule(ResponsePropertyMaxLengthUnsetId, ERR, ResponsePropertyMaxLengthUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.maxLength:unset"),
		// ResponsePropertyMinDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minimum:decrease"),
		newBackwardCompatibilityRule(ResponsePropertyMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minimum:decrease"),
		newBackwardCompatibilityRule(ResponseBodyExclusiveMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.exclusiveMinimum:decrease"),
		newBackwardCompatibilityRule(ResponsePropertyExclusiveMinDecreasedId, ERR, ResponsePropertyMinDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.exclusiveMinimum:decrease"),
		// ResponsePropertyMinItemsDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinItemsDecreasedId, ERR, ResponsePropertyMinItemsDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minItems:decrease"),
		newBackwardCompatibilityRule(ResponsePropertyMinItemsDecreasedId, ERR, ResponsePropertyMinItemsDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minItems:decrease"),
		// ResponsePropertyMinItemsUnsetCheck
		newBackwardCompatibilityRule(ResponseBodyMinItemsUnsetId, ERR, ResponsePropertyMinItemsUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.minItems:unset"),
		newBackwardCompatibilityRule(ResponsePropertyMinItemsUnsetId, ERR, ResponsePropertyMinItemsUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.minItems:unset"),
		// ResponsePropertyMinLengthDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinLengthDecreasedId, ERR, ResponsePropertyMinLengthDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minLength:decrease"),
		newBackwardCompatibilityRule(ResponsePropertyMinLengthDecreasedId, ERR, ResponsePropertyMinLengthDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minLength:decrease"),
		// ResponsePropertyOneOfUpdated
		newBackwardCompatibilityRule(ResponseBodyOneOfAddedId, ERR, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.oneOf.*:add"),
		newBackwardCompatibilityRule(ResponseBodyOneOfRemovedId, INFO, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.oneOf.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyOneOfAddedId, ERR, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.oneOf.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyOneOfRemovedId, INFO, ResponsePropertyOneOfUpdated, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.oneOf.*:remove"),
		// ResponsePropertyTypeChangedCheck
		newBackwardCompatibilityRule(ResponseBodyTypeChangedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.type:add,remove", "paths.*.*.responses.*.content.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(ResponseBodyTypeGeneralizedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionGeneralize, "paths.*.*.responses.*.content.*.schema.type:add", "paths.*.*.responses.*.content.*.schema.format:unset,change"),
		newBackwardCompatibilityRule(ResponseBodyTypeSpecializedId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionSpecialize, "paths.*.*.responses.*.content.*.schema.type:remove", "paths.*.*.responses.*.content.*.schema.format:set,change"),
		newBackwardCompatibilityRule(ResponseBodyTypeCompatibleId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.type:add,remove", "paths.*.*.responses.*.content.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(ResponsePropertyTypeChangedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.type:add,remove", "paths.*.*.responses.*.content.*.schema.format:set,unset,change"),
		newBackwardCompatibilityRule(ResponsePropertyTypeGeneralizedId, ERR, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionGeneralize, "paths.*.*.responses.*.content.*.schema.type:add", "paths.*.*.responses.*.content.*.schema.format:unset,change"),
		newBackwardCompatibilityRule(ResponsePropertyTypeSpecializedId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionSpecialize, "paths.*.*.responses.*.content.*.schema.type:remove", "paths.*.*.responses.*.content.*.schema.format:set,change"),
		newBackwardCompatibilityRule(ResponsePropertyTypeCompatibleId, INFO, ResponsePropertyTypeChangedCheck, DirectionResponse, AreaSchema, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.type:add,remove", "paths.*.*.responses.*.content.*.schema.format:set,unset,change"),
		// ResponseRequiredPropertyUpdatedCheck
		newBackwardCompatibilityRule(ResponseRequiredPropertyRemovedId, ERR, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.schema.properties.*:remove"),
		newBackwardCompatibilityRule(ResponseRequiredWriteOnlyPropertyRemovedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.schema.properties.*:remove"),
		newBackwardCompatibilityRule(ResponseRequiredPropertyAddedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.schema.properties.*:add"),
		newBackwardCompatibilityRule(ResponseRequiredWriteOnlyPropertyAddedId, INFO, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.schema.properties.*:add"),
		newBackwardCompatibilityRule(ResponseBodyWrappedInOneOfId, ERR, ResponseRequiredPropertyUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange, "paths.*.*.responses.*.content.*.schema.oneOf.*:add"),
		// ResponseRequiredPropertyWriteOnlyReadOnlyCheck
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameNonWriteOnlyId, WARN, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.writeOnly:unset"),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameWriteOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.writeOnly:set"),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameReadOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.readOnly:set"),
		newBackwardCompatibilityRule(ResponseRequiredPropertyBecameNonReadOnlyId, INFO, ResponseRequiredPropertyWriteOnlyReadOnlyCheck, DirectionResponse, AreaSchema, KindMutability, ActionChange, "paths.*.*.responses.*.content.*.schema.readOnly:unset"),
		// ResponseSuccessStatusUpdatedCheck
		newBackwardCompatibilityRule(ResponseSuccessStatusRemovedId, ERR, ResponseSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove, "paths.*.*.responses.*:remove"),
		newBackwardCompatibilityRule(ResponseSuccessStatusAddedId, INFO, ResponseSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd, "paths.*.*.responses.*:add"),
		// ResponseNonSuccessStatusUpdatedCheck
		newBackwardCompatibilityRule(ResponseNonSuccessStatusRemovedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionRemove, "paths.*.*.responses.*:remove"),
		newBackwardCompatibilityRule(ResponseNonSuccessStatusAddedId, INFO, ResponseNonSuccessStatusUpdatedCheck, DirectionResponse, AreaResponses, KindExistence, ActionAdd, "paths.*.*.responses.*:add"),
		// APIOperationIdUpdatedCheck
		newBackwardCompatibilityRule(APIOperationIdRemovedId, INFO, APIOperationIdUpdatedCheck, DirectionNone, AreaPaths, KindExistence, ActionRemove, "paths.*.*.operationId:unset,change"),
		newBackwardCompatibilityRule(APIOperationIdAddId, INFO, APIOperationIdUpdatedCheck, DirectionNone, AreaPaths, KindExistence, ActionAdd, "paths.*.*.operationId:set,change"),
		// APITagUpdatedCheck
		newBackwardCompatibilityRule(APITagRemovedId, INFO, APITagUpdatedCheck, DirectionNone, AreaTags, KindExistence, ActionRemove, "paths.*.*.tags:remove"),
		newBackwardCompatibilityRule(APITagAddedId, INFO, APITagUpdatedCheck, DirectionNone, AreaTags, KindExistence, ActionAdd, "paths.*.*.tags:add"),
		// WebhookUpdatedCheck
		newBackwardCompatibilityRule(WebhookAddedId, INFO, WebhookUpdatedCheck, DirectionNone, AreaComponents, KindExistence, ActionAdd, "webhooks.*:add", "webhooks.*.*:add"),
		newBackwardCompatibilityRule(WebhookRemovedId, ERR, WebhookUpdatedCheck, DirectionNone, AreaComponents, KindExistence, ActionRemove, "webhooks.*:remove", "webhooks.*.*:remove"),
		// APIComponentsSchemaRemovedCheck
		newBackwardCompatibilityRule(APISchemasRemovedId, INFO, APIComponentsSchemaRemovedCheck, DirectionNone, AreaComponents, KindExistence, ActionRemove, "components.schemas.*:remove"),
		// ResponseParameterEnumValueRemovedCheck
		newBackwardCompatibilityRule(ResponsePropertyEnumValueRemovedId, INFO, ResponseParameterEnumValueRemovedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove, "paths.*.*.responses.*.content.*.schema.enum:remove"),
		// ResponseMediaTypeEnumValueRemovedCheck
		newBackwardCompatibilityRule(ResponseMediaTypeEnumValueRemovedId, INFO, ResponseMediaTypeEnumValueRemovedCheck, DirectionResponse, AreaSchema, KindValues, ActionRemove, "paths.*.*.responses.*.content.*.schema.enum:remove"),
		// RequestBodyEnumValueRemovedCheck: removing a value from a request body
		// enum rejects input a client used to send, so it is breaking, the same
		// as its request-property and request-parameter siblings.
		newBackwardCompatibilityRule(RequestBodyEnumValueRemovedId, ERR, RequestBodyEnumValueRemovedCheck, DirectionRequest, AreaSchema, KindValues, ActionRemove, "paths.*.*.requestBody.content.*.schema.enum:remove"),
		// RequestPropertyListOfTypesChangedCheck
		newBackwardCompatibilityRule(RequestBodyListOfTypesWidenedId, INFO, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionAdd, "paths.*.*.requestBody.content.*.schema.type:add"),
		newBackwardCompatibilityRule(RequestBodyListOfTypesNarrowedId, ERR, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionRemove, "paths.*.*.requestBody.content.*.schema.type:remove"),
		newBackwardCompatibilityRule(RequestPropertyListOfTypesWidenedId, INFO, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionAdd, "paths.*.*.requestBody.content.*.schema.type:add"),
		newBackwardCompatibilityRule(RequestPropertyListOfTypesNarrowedId, ERR, RequestPropertyListOfTypesChangedCheck, DirectionRequest, AreaSchema, KindType, ActionRemove, "paths.*.*.requestBody.content.*.schema.type:remove"),
		// ResponsePropertyListOfTypesChangedCheck
		newBackwardCompatibilityRule(ResponseBodyListOfTypesWidenedId, ERR, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionAdd, "paths.*.*.responses.*.content.*.schema.type:add"),
		newBackwardCompatibilityRule(ResponseBodyListOfTypesNarrowedId, INFO, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionRemove, "paths.*.*.responses.*.content.*.schema.type:remove"),
		newBackwardCompatibilityRule(ResponsePropertyListOfTypesWidenedId, ERR, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionAdd, "paths.*.*.responses.*.content.*.schema.type:add"),
		newBackwardCompatibilityRule(ResponsePropertyListOfTypesNarrowedId, INFO, ResponsePropertyListOfTypesChangedCheck, DirectionResponse, AreaSchema, KindType, ActionRemove, "paths.*.*.responses.*.content.*.schema.type:remove"),
		// RequestParameterListOfTypesChangedCheck
		newBackwardCompatibilityRule(RequestParameterListOfTypesWidenedId, INFO, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionAdd, "paths.*.*.parameters.*.schema.type:add"),
		newBackwardCompatibilityRule(RequestParameterListOfTypesNarrowedId, ERR, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionRemove, "paths.*.*.parameters.*.schema.type:remove"),
		newBackwardCompatibilityRule(RequestParameterPropertyListOfTypesWidenedId, INFO, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionAdd, "paths.*.*.parameters.*.schema.type:add"),
		newBackwardCompatibilityRule(RequestParameterPropertyListOfTypesNarrowedId, ERR, RequestParameterListOfTypesChangedCheck, DirectionRequest, AreaParameters, KindType, ActionRemove, "paths.*.*.parameters.*.schema.type:remove"),
		// RequestPropertyPrefixItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyPrefixItemsAddedId, INFO, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.prefixItems.*:add"),
		newBackwardCompatibilityRule(RequestBodyPrefixItemsRemovedId, ERR, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.prefixItems.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyPrefixItemsAddedId, INFO, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.prefixItems.*:add"),
		newBackwardCompatibilityRule(RequestPropertyPrefixItemsRemovedId, ERR, RequestPropertyPrefixItemsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.prefixItems.*:remove"),
		// ResponsePropertyPrefixItemsUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyPrefixItemsAddedId, ERR, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.prefixItems.*:add"),
		newBackwardCompatibilityRule(ResponseBodyPrefixItemsRemovedId, INFO, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.prefixItems.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyPrefixItemsAddedId, ERR, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.prefixItems.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyPrefixItemsRemovedId, INFO, ResponsePropertyPrefixItemsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.prefixItems.*:remove"),
		// RequestPropertyIfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyIfAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.if:set"),
		newBackwardCompatibilityRule(RequestBodyIfRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.if:unset"),
		newBackwardCompatibilityRule(RequestBodyThenAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.then:set"),
		newBackwardCompatibilityRule(RequestBodyThenRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.then:unset"),
		newBackwardCompatibilityRule(RequestBodyElseAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.else:set"),
		newBackwardCompatibilityRule(RequestBodyElseRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.else:unset"),
		newBackwardCompatibilityRule(RequestPropertyIfAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.if:set"),
		newBackwardCompatibilityRule(RequestPropertyIfRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.if:unset"),
		newBackwardCompatibilityRule(RequestPropertyThenAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.then:set"),
		newBackwardCompatibilityRule(RequestPropertyThenRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.then:unset"),
		newBackwardCompatibilityRule(RequestPropertyElseAddedId, ERR, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.else:set"),
		newBackwardCompatibilityRule(RequestPropertyElseRemovedId, INFO, RequestPropertyIfUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.else:unset"),
		// ResponsePropertyIfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyIfAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.if:set"),
		newBackwardCompatibilityRule(ResponseBodyIfRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.if:unset"),
		newBackwardCompatibilityRule(ResponseBodyThenAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.then:set"),
		newBackwardCompatibilityRule(ResponseBodyThenRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.then:unset"),
		newBackwardCompatibilityRule(ResponseBodyElseAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.else:set"),
		newBackwardCompatibilityRule(ResponseBodyElseRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.else:unset"),
		newBackwardCompatibilityRule(ResponsePropertyIfAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.if:set"),
		newBackwardCompatibilityRule(ResponsePropertyIfRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.if:unset"),
		newBackwardCompatibilityRule(ResponsePropertyThenAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.then:set"),
		newBackwardCompatibilityRule(ResponsePropertyThenRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.then:unset"),
		newBackwardCompatibilityRule(ResponsePropertyElseAddedId, INFO, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.else:set"),
		newBackwardCompatibilityRule(ResponsePropertyElseRemovedId, ERR, ResponsePropertyIfUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.else:unset"),
		// RequestPropertyContainsUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyContainsAddedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.contains:set"),
		newBackwardCompatibilityRule(RequestBodyContainsRemovedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.contains:unset"),
		newBackwardCompatibilityRule(RequestBodyMinContainsIncreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minContains:increase"),
		newBackwardCompatibilityRule(RequestBodyMinContainsDecreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.minContains:decrease"),
		newBackwardCompatibilityRule(RequestBodyMaxContainsIncreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxContains:increase"),
		newBackwardCompatibilityRule(RequestBodyMaxContainsDecreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxContains:decrease"),
		newBackwardCompatibilityRule(RequestPropertyContainsAddedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.contains:set"),
		newBackwardCompatibilityRule(RequestPropertyContainsRemovedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.contains:unset"),
		newBackwardCompatibilityRule(RequestPropertyMinContainsIncreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minContains:increase"),
		newBackwardCompatibilityRule(RequestPropertyMinContainsDecreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.minContains:decrease"),
		newBackwardCompatibilityRule(RequestPropertyMaxContainsIncreasedId, INFO, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxContains:increase"),
		newBackwardCompatibilityRule(RequestPropertyMaxContainsDecreasedId, ERR, RequestPropertyContainsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxContains:decrease"),
		// ResponsePropertyContainsUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyContainsAddedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.contains:set"),
		newBackwardCompatibilityRule(ResponseBodyContainsRemovedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.contains:unset"),
		newBackwardCompatibilityRule(ResponseBodyMinContainsIncreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.minContains:increase"),
		newBackwardCompatibilityRule(ResponseBodyMinContainsDecreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minContains:decrease"),
		newBackwardCompatibilityRule(ResponseBodyMaxContainsIncreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxContains:increase"),
		newBackwardCompatibilityRule(ResponseBodyMaxContainsDecreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.maxContains:decrease"),
		newBackwardCompatibilityRule(ResponsePropertyContainsAddedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.contains:set"),
		newBackwardCompatibilityRule(ResponsePropertyContainsRemovedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.contains:unset"),
		newBackwardCompatibilityRule(ResponsePropertyMinContainsIncreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.minContains:increase"),
		newBackwardCompatibilityRule(ResponsePropertyMinContainsDecreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minContains:decrease"),
		newBackwardCompatibilityRule(ResponsePropertyMaxContainsIncreasedId, ERR, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxContains:increase"),
		newBackwardCompatibilityRule(ResponsePropertyMaxContainsDecreasedId, INFO, ResponsePropertyContainsUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.maxContains:decrease"),
		// RequestPropertyDependentRequiredChangedCheck
		newBackwardCompatibilityRule(RequestBodyDependentRequiredAddedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.dependentRequired.*:add"),
		newBackwardCompatibilityRule(RequestBodyDependentRequiredRemovedId, INFO, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.dependentRequired.*:remove"),
		newBackwardCompatibilityRule(RequestBodyDependentRequiredChangedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange, "paths.*.*.requestBody.content.*.schema.dependentRequired.*:add,remove"),
		newBackwardCompatibilityRule(RequestPropertyDependentRequiredAddedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.dependentRequired.*:add"),
		newBackwardCompatibilityRule(RequestPropertyDependentRequiredRemovedId, INFO, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.dependentRequired.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyDependentRequiredChangedId, ERR, RequestPropertyDependentRequiredChangedCheck, DirectionRequest, AreaSchema, KindStructure, ActionChange, "paths.*.*.requestBody.content.*.schema.dependentRequired.*:add,remove"),
		// ResponsePropertyDependentRequiredChangedCheck
		newBackwardCompatibilityRule(ResponseBodyDependentRequiredAddedId, INFO, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.dependentRequired.*:add"),
		newBackwardCompatibilityRule(ResponseBodyDependentRequiredRemovedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.dependentRequired.*:remove"),
		newBackwardCompatibilityRule(ResponseBodyDependentRequiredChangedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange, "paths.*.*.responses.*.content.*.schema.dependentRequired.*:add,remove"),
		newBackwardCompatibilityRule(ResponsePropertyDependentRequiredAddedId, INFO, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.dependentRequired.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyDependentRequiredRemovedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.dependentRequired.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyDependentRequiredChangedId, ERR, ResponsePropertyDependentRequiredChangedCheck, DirectionResponse, AreaSchema, KindStructure, ActionChange, "paths.*.*.responses.*.content.*.schema.dependentRequired.*:add,remove"),
		// RequestPropertyDependentSchemasUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyDependentSchemaAddedId, ERR, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.dependentSchemas.*:add"),
		newBackwardCompatibilityRule(RequestBodyDependentSchemaRemovedId, INFO, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.dependentSchemas.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyDependentSchemaAddedId, ERR, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.dependentSchemas.*:add"),
		newBackwardCompatibilityRule(RequestPropertyDependentSchemaRemovedId, INFO, RequestPropertyDependentSchemasUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.dependentSchemas.*:remove"),
		// ResponsePropertyDependentSchemasUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyDependentSchemaAddedId, INFO, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.dependentSchemas.*:add"),
		newBackwardCompatibilityRule(ResponseBodyDependentSchemaRemovedId, ERR, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.dependentSchemas.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyDependentSchemaAddedId, INFO, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.dependentSchemas.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyDependentSchemaRemovedId, ERR, ResponsePropertyDependentSchemasUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.dependentSchemas.*:remove"),
		// RequestPropertyPatternPropertiesUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyPatternPropertyAddedId, ERR, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionAdd, "paths.*.*.requestBody.content.*.schema.patternProperties.*:add"),
		newBackwardCompatibilityRule(RequestBodyPatternPropertyRemovedId, INFO, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.requestBody.content.*.schema.patternProperties.*:remove"),
		newBackwardCompatibilityRule(RequestPropertyPatternPropertyAddedId, ERR, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionAdd, "paths.*.*.requestBody.content.*.schema.patternProperties.*:add"),
		newBackwardCompatibilityRule(RequestPropertyPatternPropertyRemovedId, INFO, RequestPropertyPatternPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.requestBody.content.*.schema.patternProperties.*:remove"),
		// ResponsePropertyPatternPropertiesUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyPatternPropertyAddedId, INFO, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionAdd, "paths.*.*.responses.*.content.*.schema.patternProperties.*:add"),
		newBackwardCompatibilityRule(ResponseBodyPatternPropertyRemovedId, ERR, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.patternProperties.*:remove"),
		newBackwardCompatibilityRule(ResponsePropertyPatternPropertyAddedId, INFO, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionAdd, "paths.*.*.responses.*.content.*.schema.patternProperties.*:add"),
		newBackwardCompatibilityRule(ResponsePropertyPatternPropertyRemovedId, ERR, ResponsePropertyPatternPropertiesUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.patternProperties.*:remove"),
		// RequestPropertyPropertyNamesUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyPropertyNamesAddedId, ERR, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.propertyNames:set"),
		newBackwardCompatibilityRule(RequestBodyPropertyNamesRemovedId, INFO, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.propertyNames:unset"),
		newBackwardCompatibilityRule(RequestPropertyPropertyNamesAddedId, ERR, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.propertyNames:set"),
		newBackwardCompatibilityRule(RequestPropertyPropertyNamesRemovedId, INFO, RequestPropertyPropertyNamesUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.propertyNames:unset"),
		// ResponsePropertyPropertyNamesUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyPropertyNamesAddedId, INFO, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.propertyNames:set"),
		newBackwardCompatibilityRule(ResponseBodyPropertyNamesRemovedId, ERR, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.propertyNames:unset"),
		newBackwardCompatibilityRule(ResponsePropertyPropertyNamesAddedId, INFO, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.propertyNames:set"),
		newBackwardCompatibilityRule(ResponsePropertyPropertyNamesRemovedId, ERR, ResponsePropertyPropertyNamesUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.propertyNames:unset"),
		// RequestPropertyUnevaluatedUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyUnevaluatedItemsAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.unevaluatedItems:set"),
		newBackwardCompatibilityRule(RequestBodyUnevaluatedItemsRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.unevaluatedItems:unset"),
		newBackwardCompatibilityRule(RequestBodyUnevaluatedPropertiesAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.unevaluatedProperties:set"),
		newBackwardCompatibilityRule(RequestBodyUnevaluatedPropertiesRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.unevaluatedProperties:unset"),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedItemsAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.unevaluatedItems:set"),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedItemsRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.unevaluatedItems:unset"),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedPropertiesAddedId, ERR, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionAdd, "paths.*.*.requestBody.content.*.schema.unevaluatedProperties:set"),
		newBackwardCompatibilityRule(RequestPropertyUnevaluatedPropertiesRemovedId, INFO, RequestPropertyUnevaluatedUpdatedCheck, DirectionRequest, AreaSchema, KindStructure, ActionRemove, "paths.*.*.requestBody.content.*.schema.unevaluatedProperties:unset"),
		// ResponsePropertyUnevaluatedUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedItemsAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.unevaluatedItems:set"),
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedItemsRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.unevaluatedItems:unset"),
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedPropertiesAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.unevaluatedProperties:set"),
		newBackwardCompatibilityRule(ResponseBodyUnevaluatedPropertiesRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.unevaluatedProperties:unset"),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedItemsAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.unevaluatedItems:set"),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedItemsRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.unevaluatedItems:unset"),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedPropertiesAddedId, INFO, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionAdd, "paths.*.*.responses.*.content.*.schema.unevaluatedProperties:set"),
		newBackwardCompatibilityRule(ResponsePropertyUnevaluatedPropertiesRemovedId, ERR, ResponsePropertyUnevaluatedUpdatedCheck, DirectionResponse, AreaSchema, KindStructure, ActionRemove, "paths.*.*.responses.*.content.*.schema.unevaluatedProperties:unset"),
		// RequestPropertyContentUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyContentSchemaAddedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*.schema.contentSchema:set"),
		newBackwardCompatibilityRule(RequestBodyContentSchemaRemovedId, INFO, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove, "paths.*.*.requestBody.content.*.schema.contentSchema:unset"),
		newBackwardCompatibilityRule(RequestBodyContentMediaTypeChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaRequestBody, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.contentMediaType:set,unset,change"),
		newBackwardCompatibilityRule(RequestBodyContentEncodingChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaRequestBody, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.contentEncoding:set,unset,change"),
		newBackwardCompatibilityRule(RequestPropertyContentSchemaAddedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionAdd, "paths.*.*.requestBody.content.*.schema.contentSchema:set"),
		newBackwardCompatibilityRule(RequestPropertyContentSchemaRemovedId, INFO, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindExistence, ActionRemove, "paths.*.*.requestBody.content.*.schema.contentSchema:unset"),
		newBackwardCompatibilityRule(RequestPropertyContentMediaTypeChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.contentMediaType:set,unset,change"),
		newBackwardCompatibilityRule(RequestPropertyContentEncodingChangedId, ERR, RequestPropertyContentUpdatedCheck, DirectionRequest, AreaSchema, KindType, ActionChange, "paths.*.*.requestBody.content.*.schema.contentEncoding:set,unset,change"),
		// ResponsePropertyContentUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyContentSchemaAddedId, INFO, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.schema.contentSchema:set"),
		newBackwardCompatibilityRule(ResponseBodyContentSchemaRemovedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.schema.contentSchema:unset"),
		newBackwardCompatibilityRule(ResponseBodyContentMediaTypeChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.contentMediaType:set,unset,change"),
		newBackwardCompatibilityRule(ResponseBodyContentEncodingChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaResponses, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.contentEncoding:set,unset,change"),
		newBackwardCompatibilityRule(ResponsePropertyContentSchemaAddedId, INFO, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionAdd, "paths.*.*.responses.*.content.*.schema.contentSchema:set"),
		newBackwardCompatibilityRule(ResponsePropertyContentSchemaRemovedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindExistence, ActionRemove, "paths.*.*.responses.*.content.*.schema.contentSchema:unset"),
		newBackwardCompatibilityRule(ResponsePropertyContentMediaTypeChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.contentMediaType:set,unset,change"),
		newBackwardCompatibilityRule(ResponsePropertyContentEncodingChangedId, ERR, ResponsePropertyContentUpdatedCheck, DirectionResponse, AreaSchema, KindType, ActionChange, "paths.*.*.responses.*.content.*.schema.contentEncoding:set,unset,change"),
		// RequestPropertyMaxItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMaxItemsDecreasedId, ERR, RequestPropertyMaxItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxItems:decrease"),
		newBackwardCompatibilityRule(RequestBodyMaxItemsIncreasedId, INFO, RequestPropertyMaxItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxItems:increase"),
		newBackwardCompatibilityRule(RequestPropertyMaxItemsDecreasedId, ERR, RequestPropertyMaxItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxItems:decrease"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxItemsDecreasedId, INFO, RequestPropertyMaxItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxItems:decrease"),
		newBackwardCompatibilityRule(RequestPropertyMaxItemsIncreasedId, INFO, RequestPropertyMaxItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxItems:increase"),
		// RequestPropertyMaxItemsSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxItemsSetId, WARN, RequestPropertyMaxItemsSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maxItems:set"),
		newBackwardCompatibilityRule(RequestPropertyMaxItemsSetId, WARN, RequestPropertyMaxItemsSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maxItems:set"),
		// ResponsePropertyMaxItemsIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxItemsIncreasedId, ERR, ResponsePropertyMaxItemsIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxItems:increase"),
		newBackwardCompatibilityRule(ResponsePropertyMaxItemsIncreasedId, ERR, ResponsePropertyMaxItemsIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxItems:increase"),
		// RequestPropertyMaxPropertiesUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMaxPropertiesDecreasedId, ERR, RequestPropertyMaxPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxProperties:decrease"),
		newBackwardCompatibilityRule(RequestBodyMaxPropertiesIncreasedId, INFO, RequestPropertyMaxPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxProperties:increase"),
		newBackwardCompatibilityRule(RequestPropertyMaxPropertiesDecreasedId, ERR, RequestPropertyMaxPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxProperties:decrease"),
		newBackwardCompatibilityRule(RequestReadOnlyPropertyMaxPropertiesDecreasedId, INFO, RequestPropertyMaxPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.requestBody.content.*.schema.maxProperties:decrease"),
		newBackwardCompatibilityRule(RequestPropertyMaxPropertiesIncreasedId, INFO, RequestPropertyMaxPropertiesUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.maxProperties:increase"),
		// RequestPropertyMaxPropertiesSetCheck
		newBackwardCompatibilityRule(RequestBodyMaxPropertiesSetId, WARN, RequestPropertyMaxPropertiesSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maxProperties:set"),
		newBackwardCompatibilityRule(RequestPropertyMaxPropertiesSetId, WARN, RequestPropertyMaxPropertiesSetCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.maxProperties:set"),
		// ResponsePropertyMaxPropertiesIncreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMaxPropertiesIncreasedId, ERR, ResponsePropertyMaxPropertiesIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxProperties:increase"),
		newBackwardCompatibilityRule(ResponsePropertyMaxPropertiesIncreasedId, ERR, ResponsePropertyMaxPropertiesIncreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.responses.*.content.*.schema.maxProperties:increase"),
		// RequestPropertyMinPropertiesIncreasedCheck
		newBackwardCompatibilityRule(RequestBodyMinPropertiesIncreasedId, ERR, RequestPropertyMinPropertiesIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minProperties:increase"),
		newBackwardCompatibilityRule(RequestPropertyMinPropertiesIncreasedId, ERR, RequestPropertyMinPropertiesIncreasedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionIncrease, "paths.*.*.requestBody.content.*.schema.minProperties:increase"),
		// ResponsePropertyMinPropertiesDecreasedCheck
		newBackwardCompatibilityRule(ResponseBodyMinPropertiesDecreasedId, ERR, ResponsePropertyMinPropertiesDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minProperties:decrease"),
		newBackwardCompatibilityRule(ResponsePropertyMinPropertiesDecreasedId, ERR, ResponsePropertyMinPropertiesDecreasedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionDecrease, "paths.*.*.responses.*.content.*.schema.minProperties:decrease"),
		// RequestPropertyUniqueItemsUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyUniqueItemsSetId, ERR, RequestPropertyUniqueItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.uniqueItems:set"),
		newBackwardCompatibilityRule(RequestPropertyUniqueItemsSetId, ERR, RequestPropertyUniqueItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.uniqueItems:set"),
		newBackwardCompatibilityRule(RequestBodyUniqueItemsUnsetId, INFO, RequestPropertyUniqueItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.requestBody.content.*.schema.uniqueItems:unset"),
		newBackwardCompatibilityRule(RequestPropertyUniqueItemsUnsetId, INFO, RequestPropertyUniqueItemsUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.requestBody.content.*.schema.uniqueItems:unset"),
		// ResponsePropertyUniqueItemsUnsetCheck
		newBackwardCompatibilityRule(ResponseBodyUniqueItemsUnsetId, ERR, ResponsePropertyUniqueItemsUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.uniqueItems:unset"),
		newBackwardCompatibilityRule(ResponsePropertyUniqueItemsUnsetId, ERR, ResponsePropertyUniqueItemsUnsetCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.uniqueItems:unset"),
		// RequestPropertyMultipleOfUpdatedCheck
		newBackwardCompatibilityRule(RequestBodyMultipleOfSetId, WARN, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.multipleOf:set"),
		newBackwardCompatibilityRule(RequestPropertyMultipleOfSetId, WARN, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionSet, "paths.*.*.requestBody.content.*.schema.multipleOf:set"),
		newBackwardCompatibilityRule(RequestBodyMultipleOfUnsetId, INFO, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.requestBody.content.*.schema.multipleOf:unset"),
		newBackwardCompatibilityRule(RequestPropertyMultipleOfUnsetId, INFO, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.requestBody.content.*.schema.multipleOf:unset"),
		newBackwardCompatibilityRule(RequestBodyMultipleOfChangedId, ERR, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionChange, "paths.*.*.requestBody.content.*.schema.multipleOf:increase,decrease"),
		newBackwardCompatibilityRule(RequestPropertyMultipleOfChangedId, ERR, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionChange, "paths.*.*.requestBody.content.*.schema.multipleOf:increase,decrease"),
		newBackwardCompatibilityRule(RequestBodyMultipleOfGeneralizedId, INFO, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionGeneralize, "paths.*.*.requestBody.content.*.schema.multipleOf:increase,decrease"),
		newBackwardCompatibilityRule(RequestPropertyMultipleOfGeneralizedId, INFO, RequestPropertyMultipleOfUpdatedCheck, DirectionRequest, AreaSchema, KindConstraints, ActionGeneralize, "paths.*.*.requestBody.content.*.schema.multipleOf:increase,decrease"),
		// ResponsePropertyMultipleOfUpdatedCheck
		newBackwardCompatibilityRule(ResponseBodyMultipleOfUnsetId, ERR, ResponsePropertyMultipleOfUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.multipleOf:unset"),
		newBackwardCompatibilityRule(ResponsePropertyMultipleOfUnsetId, ERR, ResponsePropertyMultipleOfUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionRemove, "paths.*.*.responses.*.content.*.schema.multipleOf:unset"),
		newBackwardCompatibilityRule(ResponseBodyMultipleOfChangedId, ERR, ResponsePropertyMultipleOfUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionChange, "paths.*.*.responses.*.content.*.schema.multipleOf:increase,decrease"),
		newBackwardCompatibilityRule(ResponsePropertyMultipleOfChangedId, ERR, ResponsePropertyMultipleOfUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionChange, "paths.*.*.responses.*.content.*.schema.multipleOf:increase,decrease"),
		newBackwardCompatibilityRule(ResponseBodyMultipleOfSpecializedId, INFO, ResponsePropertyMultipleOfUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionSpecialize, "paths.*.*.responses.*.content.*.schema.multipleOf:increase,decrease"),
		newBackwardCompatibilityRule(ResponsePropertyMultipleOfSpecializedId, INFO, ResponsePropertyMultipleOfUpdatedCheck, DirectionResponse, AreaSchema, KindConstraints, ActionSpecialize, "paths.*.*.responses.*.content.*.schema.multipleOf:increase,decrease"),
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
