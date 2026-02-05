package lint

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/utils"
)

func checkRequireProperties(schema *openapi3.Schema, s *state) *Error {
	if schema == nil {
		return nil
	}

	requiredProps := utils.StringSetFromSlice(schema.Required)
	props := utils.StringSet{}
	for name := range schema.Properties {
		props.Add(name)
	}

	if extraRequiredProps := requiredProps.Minus(props); !extraRequiredProps.Empty() {
		return &Error{
			Id:     "extra_required_props",
			Level:  LEVEL_ERROR,
			Text:   fmt.Sprintf("none-existing properties %v defined as required", extraRequiredProps.ToStringList()),
			Source: s.source,
		}
	}

	return nil
}
