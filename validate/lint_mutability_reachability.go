package validate

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// ReadOnlyOnlyInRequestsID flags a readOnly property whose schema is
// reachable only from request bodies and parameters. A readOnly property
// never appears in a request, so in a schema no response uses it declares a
// property that can never appear anywhere: either the flag or the schema's
// usage is wrong. WriteOnlyOnlyInResponsesID is the mirror for writeOnly
// properties in schemas no request uses. Valid OpenAPI, oasdiff-native
// SHOULD-level lints; the flags also drive the changelog's read-only and
// write-only verdicts, so a wrong flag silently mutes breaking-change
// detection for the property.
const (
	ReadOnlyOnlyInRequestsID   = "read-only-property-only-in-requests"
	WriteOnlyOnlyInResponsesID = "write-only-property-only-in-responses"
)

// lintMutabilityReachability reports, at WARN, every readOnly property
// reachable only from the request side and every writeOnly property
// reachable only from the response side. Reachability is computed from the
// operations (including webhooks and callbacks) by schema identity, which
// resolved $refs share, so a component schema used by both sides clears
// its properties wherever they are reported.
func lintMutabilityReachability(spec *openapi3.T, source string) formatters.Findings {
	requestReachable := map[*openapi3.Schema]bool{}
	responseReachable := map[*openapi3.Schema]bool{}

	if spec.Paths != nil {
		for _, pathItem := range spec.Paths.Map() {
			markPathItem(pathItem, requestReachable, responseReachable)
		}
	}
	for _, pathItem := range spec.Webhooks {
		markPathItem(pathItem, requestReachable, responseReachable)
	}

	var findings formatters.Findings
	_ = spec.WalkSchemas(func(jsonPointer string, ref *openapi3.SchemaRef) error {
		for name, prop := range ref.Value.Properties {
			p := prop.Value
			if p == nil {
				continue
			}
			if p.ReadOnly && requestReachable[p] && !responseReachable[p] {
				findings = append(findings, newMutabilityReachabilityFinding(
					ReadOnlyOnlyInRequestsID,
					fmt.Sprintf("readOnly property %q is only reachable from the request side, where a readOnly property never appears; either the flag or the schema's usage is wrong", name),
					jsonPointer+"/properties/"+name, name, p, "readOnly", source))
			}
			if p.WriteOnly && responseReachable[p] && !requestReachable[p] {
				findings = append(findings, newMutabilityReachabilityFinding(
					WriteOnlyOnlyInResponsesID,
					fmt.Sprintf("writeOnly property %q is only reachable from the response side, where a writeOnly property never appears; either the flag or the schema's usage is wrong", name),
					jsonPointer+"/properties/"+name, name, p, "writeOnly", source))
			}
		}
		return nil
	})
	return findings
}

func markPathItem(pathItem *openapi3.PathItem, request, response map[*openapi3.Schema]bool) {
	if pathItem == nil {
		return
	}
	for _, param := range pathItem.Parameters {
		markParameter(param, request)
	}
	for _, op := range pathItem.Operations() {
		markOperation(op, request, response)
	}
}

func markOperation(op *openapi3.Operation, request, response map[*openapi3.Schema]bool) {
	if op == nil {
		return
	}
	for _, param := range op.Parameters {
		markParameter(param, request)
	}
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		markContent(op.RequestBody.Value.Content, request)
	}
	if op.Responses != nil {
		for _, resp := range op.Responses.Map() {
			if resp == nil || resp.Value == nil {
				continue
			}
			markContent(resp.Value.Content, response)
			for _, header := range resp.Value.Headers {
				if header == nil || header.Value == nil {
					continue
				}
				markSchema(header.Value.Schema, response)
				markContent(header.Value.Content, response)
			}
		}
	}
	for _, callback := range op.Callbacks {
		if callback == nil || callback.Value == nil {
			continue
		}
		for _, pathItem := range callback.Value.Map() {
			markPathItem(pathItem, request, response)
		}
	}
}

func markParameter(param *openapi3.ParameterRef, request map[*openapi3.Schema]bool) {
	if param == nil || param.Value == nil {
		return
	}
	markSchema(param.Value.Schema, request)
	markContent(param.Value.Content, request)
}

func markContent(content openapi3.Content, set map[*openapi3.Schema]bool) {
	for _, mediaType := range content {
		if mediaType == nil {
			continue
		}
		markSchema(mediaType.Schema, set)
	}
}

// markSchema marks the schema and every sub-schema reachable from it. The
// set doubles as the cycle guard: a marked schema's subtree is already
// marked.
func markSchema(ref *openapi3.SchemaRef, set map[*openapi3.Schema]bool) {
	if ref == nil || ref.Value == nil || set[ref.Value] {
		return
	}
	s := ref.Value
	set[s] = true

	for _, prop := range s.Properties {
		markSchema(prop, set)
	}
	for _, sub := range s.AllOf {
		markSchema(sub, set)
	}
	for _, sub := range s.AnyOf {
		markSchema(sub, set)
	}
	for _, sub := range s.OneOf {
		markSchema(sub, set)
	}
	for _, sub := range s.PrefixItems {
		markSchema(sub, set)
	}
	for _, sub := range s.PatternProperties {
		markSchema(sub, set)
	}
	for _, sub := range s.DependentSchemas {
		markSchema(sub, set)
	}
	markSchema(s.Items, set)
	markSchema(s.Not, set)
	markSchema(s.Contains, set)
	markSchema(s.PropertyNames, set)
	markSchema(s.If, set)
	markSchema(s.Then, set)
	markSchema(s.Else, set)
	markSchema(s.ContentSchema, set)
	markSchema(s.AdditionalProperties.Schema, set)
	markSchema(s.UnevaluatedItems.Schema, set)
	markSchema(s.UnevaluatedProperties.Schema, set)
}

func newMutabilityReachabilityFinding(id, text, section, name string, prop *openapi3.Schema, field, source string) formatters.Finding {
	line, column := schemaFieldLocation(prop, field)
	f := formatters.Finding{
		Id:      id,
		Text:    text,
		Level:   RuleLevel(id),
		Section: section,
		Source: formatters.Source{
			File:   source,
			Line:   line,
			Column: column,
		},
	}
	f.Fingerprint = checker.ComputeFingerprint(f.Id, "", section, []any{name})
	return f
}
