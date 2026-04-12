package allof

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// MergeSpec merges all instances in allOf in place
func MergeSpec(spec *openapi3.T) (*openapi3.T, error) {

	if err := mergeComponents(spec.Components); err != nil {
		return spec, err
	}

	for _, v := range spec.Paths.Map() {
		if v == nil {
			continue
		}
		if _, err := mergePathItem(v); err != nil {
			return spec, err
		}
	}
	return spec, nil
}

func mergeComponents(components *openapi3.Components) error {
	if components == nil {
		return nil
	}

	var err error

	if components.Schemas, err = mergeSchemas(components.Schemas); err != nil {
		return err
	}

	if components.Parameters, err = mergeParametersMap(components.Parameters); err != nil {
		return err
	}

	if components.Headers, err = mergeHeaders(components.Headers); err != nil {
		return err
	}

	if components.RequestBodies, err = mergeRequestBodies(components.RequestBodies); err != nil {
		return err
	}

	if components.Responses, err = mergeResponseBodies(components.Responses); err != nil {
		return err
	}

	if components.Callbacks, err = mergeCallbacks(components.Callbacks); err != nil {
		return err
	}

	return nil
}

func mergeOperation(operation *openapi3.Operation) (*openapi3.Operation, error) {
	parameteres, err := mergeParameters(operation.Parameters)
	if err != nil {
		return operation, err
	}
	operation.Parameters = parameteres
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		content, err := mergeContent(operation.RequestBody.Value.Content)
		if err != nil {
			return operation, err
		}
		operation.RequestBody.Value.Content = content
	}
	responses, err := mergeResponses(operation.Responses)
	if err != nil {
		return operation, err
	}
	operation.Responses = responses
	callbacks, err := mergeCallbacks(operation.Callbacks)
	if err != nil {
		return operation, err
	}
	operation.Callbacks = callbacks
	return operation, nil
}

func mergePathItem(pathItem *openapi3.PathItem) (*openapi3.PathItem, error) {
	operations := []*openapi3.Operation{
		pathItem.Connect, pathItem.Delete, pathItem.Get, pathItem.Head,
		pathItem.Options, pathItem.Patch, pathItem.Post, pathItem.Put, pathItem.Trace,
	}

	for _, op := range operations {
		if op != nil {
			if _, err := mergeOperation(op); err != nil {
				return pathItem, err
			}
		}
	}

	parameters, err := mergeParameters(pathItem.Parameters)
	if err != nil {
		return pathItem, err
	}
	pathItem.Parameters = parameters
	return pathItem, nil
}

func mergeCallbacks(callbacks openapi3.Callbacks) (openapi3.Callbacks, error) {
	for _, v := range callbacks {
		for _, pathItem := range v.Value.Map() {
			if _, err := mergePathItem(pathItem); err != nil {
				return callbacks, err
			}
		}
	}
	return callbacks, nil
}

func mergeSchemas(schemas openapi3.Schemas) (openapi3.Schemas, error) {
	for _, s := range schemas {
		if s == nil || s.Value == nil {
			continue
		}
		m, err := Merge(*s)
		if err != nil {
			return schemas, err
		}
		s.Value = m
	}
	return schemas, nil
}

func mergeResponseBodies(responseBodies openapi3.ResponseBodies) (openapi3.ResponseBodies, error) {
	for _, v := range responseBodies {
		if v == nil || v.Value == nil {
			continue
		}
		content, err := mergeContent(v.Value.Content)
		if err != nil {
			return responseBodies, err
		}
		v.Value.Content = content
		if _, err := mergeHeaders(v.Value.Headers); err != nil {
			return responseBodies, err
		}
	}
	return responseBodies, nil
}

func mergeResponses(responses *openapi3.Responses) (*openapi3.Responses, error) {
	for _, v := range responses.Map() {
		if v == nil || v.Value == nil {
			continue
		}
		content, err := mergeContent(v.Value.Content)
		if err != nil {
			return responses, err
		}
		v.Value.Content = content
		if _, err := mergeHeaders(v.Value.Headers); err != nil {
			return responses, err
		}
	}
	return responses, nil
}

func mergeRequestBodies(rb openapi3.RequestBodies) (openapi3.RequestBodies, error) {
	for _, v := range rb {
		if v == nil || v.Value == nil {
			continue
		}
		content, err := mergeContent(v.Value.Content)
		if err != nil {
			return rb, err
		}
		v.Value.Content = content
	}
	return rb, nil
}

func mergeParameters(parameters openapi3.Parameters) (openapi3.Parameters, error) {
	for _, v := range parameters {
		if v == nil || v.Value == nil {
			continue
		}
		if _, err := mergeParameter(v.Value); err != nil {
			return parameters, err
		}
	}
	return parameters, nil
}

func mergeParametersMap(parameters openapi3.ParametersMap) (openapi3.ParametersMap, error) {
	for _, v := range parameters {
		if v == nil || v.Value == nil {
			continue
		}
		if _, err := mergeParameter(v.Value); err != nil {
			return parameters, err
		}
	}
	return parameters, nil
}

func mergeParameter(p *openapi3.Parameter) (*openapi3.Parameter, error) {
	if p.Schema == nil || p.Schema.Value == nil {
		return p, nil
	}
	m, err := Merge(*p.Schema)
	if err != nil {
		return p, err
	}
	p.Schema.Value = m
	content, err := mergeContent(p.Content)
	if err != nil {
		return p, err
	}
	p.Content = content
	return p, nil
}

func mergeContent(content openapi3.Content) (openapi3.Content, error) {
	for _, mediaType := range content {
		if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
			continue
		}
		m, err := Merge(*mediaType.Schema)
		if err != nil {
			return content, err
		}
		mediaType.Schema.Value = m
		for _, encoding := range mediaType.Encoding {
			if encoding == nil {
				continue
			}
			headers, err := mergeHeaders(encoding.Headers)
			if err != nil {
				return content, err
			}
			encoding.Headers = headers
		}
	}
	return content, nil
}

func mergeHeaders(headers openapi3.Headers) (openapi3.Headers, error) {
	for _, v := range headers {
		if v == nil || v.Value == nil {
			continue
		}
		if _, err := mergeParameter(&v.Value.Parameter); err != nil {
			return headers, err
		}
	}
	return headers, nil
}
