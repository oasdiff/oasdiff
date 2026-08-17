package rules

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
