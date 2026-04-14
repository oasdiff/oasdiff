package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestDiff_SchemaRefNil(t *testing.T) {
	loader := openapi3.NewLoader()
	s1, err := loader.LoadFromFile("../data/home-iot-api-1.yaml")
	require.NoError(t, err)

	s1.Components.Schemas["LightingSummary"].Value.Properties["zones"].Value.Items.Value.Properties["deviceId"].Value = nil
	_, err = diff.Get(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "base schema value is nil")
}

func TestDiff_MediaTypeNil(t *testing.T) {
	loader := openapi3.NewLoader()
	s1, err := loader.LoadFromFile("../data/home-iot-api-1.yaml")
	require.NoError(t, err)

	s1.Paths.Value("/devices").Post.RequestBody.Value.Content["application/json"] = nil
	_, err = diff.Get(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "media type is nil")
}

func TestDiff_EncodingNil(t *testing.T) {
	s1 := l(t, 1)

	callback := s1.Paths.Value("/subscribe").Post.Callbacks["myEvent"].Value
	require.NotNil(t, callback)

	mediaType := (*callback).Value("hi").Post.RequestBody.Value.Content["application/json"]
	require.NotNil(t, mediaType)

	mediaType.Encoding["historyMetadata"] = nil

	_, err := diff.Get(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "encoding is nil")
}

func TestDiff_PathItemNil(t *testing.T) {
	loader := openapi3.NewLoader()
	s1, err := loader.LoadFromFile("../data/home-iot-api-1.yaml")
	require.NoError(t, err)

	s1.Paths.Set("/devices", nil)
	_, err = diff.Get(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "path item is nil")
}

func TestDiff_SpecNil(t *testing.T) {
	loader := openapi3.NewLoader()
	s1, err := loader.LoadFromFile("../data/home-iot-api-1.yaml")

	require.NoError(t, err)
	_, err = diff.Get(diff.NewConfig(), nil, s1)
	require.EqualError(t, err, "spec is nil")
}

func TestDiff_ExampleNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Examples: openapi3.Examples{"test": &openapi3.ExampleRef{Value: &openapi3.Example{}}},
		},
	}
	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Examples: openapi3.Examples{"test": &openapi3.ExampleRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "example reference is nil")

	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "example reference is nil")
}

func TestDiff_ComponentSchemaNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{"test": &openapi3.SchemaRef{Value: &openapi3.Schema{}}},
		},
	}
	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{"test": &openapi3.SchemaRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "revision schema value is nil")

	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "base schema value is nil")
}

func TestDiff_ComponentSchemaDeepNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"test": &openapi3.SchemaRef{
					Value: nil,
				},
			},
		},
	}

	_, err := diff.Get(diff.NewConfig(), &s1, &s1)
	require.EqualError(t, err, "base schema value is nil")
}

func TestDiff_ComponentParameterNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Parameters: openapi3.ParametersMap{"test": &openapi3.ParameterRef{Value: &openapi3.Parameter{}}},
		},
	}
	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Parameters: openapi3.ParametersMap{"test": &openapi3.ParameterRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "parameter reference is nil")

	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "parameter reference is nil")
}

func TestDiff_ComponentHeadersNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Headers: openapi3.Headers{"test": &openapi3.HeaderRef{Value: &openapi3.Header{}}},
		},
	}
	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Headers: openapi3.Headers{"test": &openapi3.HeaderRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "header reference is nil")

	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "header reference is nil")
}

func TestDiff_ComponentRequestBodiesNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			RequestBodies: openapi3.RequestBodies{"test": &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{}}},
		},
	}
	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			RequestBodies: openapi3.RequestBodies{"test": &openapi3.RequestBodyRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "request body reference is nil")

	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "request body reference is nil")
}

func TestDiff_ComponentResponsesNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Responses: openapi3.ResponseBodies{"test": &openapi3.ResponseRef{Value: &openapi3.Response{}}},
		},
	}
	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Responses: openapi3.ResponseBodies{"test": &openapi3.ResponseRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "response reference is nil")

	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "response reference is nil")
}

func TestDiff_ComponentSecuritySchemesNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{"test": &openapi3.SecuritySchemeRef{Value: &openapi3.SecurityScheme{}}},
		},
	}

	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{"test": &openapi3.SecuritySchemeRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "security scheme reference is nil")
	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "security scheme reference is nil")
}

func TestDiff_ComponentCallbacksNil(t *testing.T) {
	s1 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Callbacks: openapi3.Callbacks{"test": &openapi3.CallbackRef{Value: &openapi3.Callback{}}},
		},
	}
	s2 := openapi3.T{
		Info: &openapi3.Info{},
		Components: &openapi3.Components{
			Callbacks: openapi3.Callbacks{"test": &openapi3.CallbackRef{}},
		},
	}
	_, err := diff.Get(diff.NewConfig(), &s1, &s2)
	require.EqualError(t, err, "callback reference is nil")

	_, err = diff.Get(diff.NewConfig(), &s2, &s1)
	require.EqualError(t, err, "callback reference is nil")
}

func TestFilterByRegex_Invalid(t *testing.T) {
	_, err := diff.Get(&diff.Config{MatchPath: "["}, l(t, 1), l(t, 2))
	require.EqualError(t, err, "failed to compile filter regex \"[\": error parsing regexp: missing closing ]: `[`")
}
