package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterStyleChangedId           = "request-parameter-style-changed"
	RequestParameterExplodeChangedId         = "request-parameter-explode-changed"
	RequestParameterBecameAllowReservedId    = "request-parameter-became-allow-reserved"
	RequestParameterBecameNotAllowReservedId = "request-parameter-became-not-allow-reserved"
)

func RequestParameterSerializationUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		base, revision := p.paramDiff.Base, p.paramDiff.Revision
		if base == nil || revision == nil {
			return
		}

		// style and explode apply to a parameter described by a schema; one
		// described by content is serialized by its media type.
		if base.Schema != nil && revision.Schema != nil {
			// The comparison is between the methods in effect, so writing out a
			// default is not a change.
			baseStyle, baseExplode := serialization(base)
			revisionStyle, revisionExplode := serialization(revision)
			if baseStyle != revisionStyle {
				baseSource, revisionSource := location.ParameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, "style")
				result = append(result, p.opInfo.NewApiChange(
					RequestParameterStyleChangedId,
					[]any{p.location, p.name, baseStyle, revisionStyle},
					"",
				).WithSources(baseSource, revisionSource))
			} else if baseExplode != revisionExplode && (!isScalarSchema(base.Schema) || !isScalarSchema(revision.Schema)) {
				baseSource, revisionSource := location.ParameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, "explode")
				result = append(result, p.opInfo.NewApiChange(
					RequestParameterExplodeChangedId,
					[]any{p.location, p.name, baseExplode, revisionExplode},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		// allowReserved applies only to query parameters.
		if p.location == openapi3.ParameterInQuery && base.AllowReserved != revision.AllowReserved {
			id := RequestParameterBecameAllowReservedId
			if !revision.AllowReserved {
				id = RequestParameterBecameNotAllowReservedId
			}
			baseSource, revisionSource := location.ParameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, "allowReserved")
			result = append(result, p.opInfo.NewApiChange(
				id,
				[]any{p.location, p.name},
				"",
			).WithSources(baseSource, revisionSource))
		}
	})
	return result
}

// serialization returns the style and explode a parameter uses. The defaults
// follow the OpenAPI specification: form for query and cookie, simple for path
// and header, and explode true only for form. kin-openapi's
// Parameter.SerializationMethod defaults explode to true for every query style.
func serialization(parameter *openapi3.Parameter) (string, bool) {
	style := parameter.Style
	if style == "" {
		switch parameter.In {
		case openapi3.ParameterInQuery, openapi3.ParameterInCookie:
			style = openapi3.SerializationForm
		default:
			style = openapi3.SerializationSimple
		}
	}
	explode := style == openapi3.SerializationForm
	if parameter.Explode != nil {
		explode = *parameter.Explode
	}
	return style, explode
}

// isScalarSchema reports whether every value the schema allows is a single
// value, for which explode makes no difference on the wire. A schema that
// does not say its type may be an array or an object.
func isScalarSchema(schemaRef *openapi3.SchemaRef) bool {
	if schemaRef == nil || schemaRef.Value == nil || schemaRef.Value.Type == nil || len(*schemaRef.Value.Type) == 0 {
		return false
	}
	schema := schemaRef.Value
	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 || len(schema.AllOf) > 0 {
		return false
	}
	for _, t := range *schema.Type {
		switch t {
		case openapi3.TypeString, openapi3.TypeNumber, openapi3.TypeInteger, openapi3.TypeBoolean, openapi3.TypeNull:
		default:
			return false
		}
	}
	return true
}
