package lint_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/lint"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func loadFrom(t *testing.T, path string) *load.SpecInfo {
	t.Helper()

	loader := openapi3.NewLoader()
	oas, err := loader.LoadFromFile(path)
	require.NoError(t, err)
	return &load.SpecInfo{Spec: oas, Url: path}
}

func TestRun(t *testing.T) {

	const source = "../data/lint/openapi.yaml"
	require.Empty(t, lint.Run(lint.DefaultConfig(), source, loadFrom(t, source)))
}

func TestRun_NilSpec(t *testing.T) {
	result := lint.Run(lint.DefaultConfig(), "test", nil)
	require.Empty(t, result)
}

func TestRun_SortByLevel(t *testing.T) {
	config := &lint.Config{
		Checks: []lint.Check{
			func(source string, spec *load.SpecInfo) []*lint.Error {
				return []*lint.Error{
					{Id: "a", Level: lint.LEVEL_WARN},
					{Id: "b", Level: lint.LEVEL_ERROR},
				}
			},
		},
	}
	const source = "../data/lint/openapi.yaml"
	result := lint.Run(config, source, loadFrom(t, source))
	require.Len(t, result, 2)
	require.Equal(t, lint.LEVEL_ERROR, result[0].Level)
	require.Equal(t, lint.LEVEL_WARN, result[1].Level)
}

func TestRun_SortBySource(t *testing.T) {
	config := &lint.Config{
		Checks: []lint.Check{
			func(source string, spec *load.SpecInfo) []*lint.Error {
				return []*lint.Error{
					{Id: "a", Level: lint.LEVEL_ERROR, Source: "z"},
					{Id: "b", Level: lint.LEVEL_ERROR, Source: "a"},
				}
			},
		},
	}
	const source = "../data/lint/openapi.yaml"
	result := lint.Run(config, source, loadFrom(t, source))
	require.Len(t, result, 2)
	require.Equal(t, "a", result[0].Source)
	require.Equal(t, "z", result[1].Source)
}

func TestRun_SortById(t *testing.T) {
	config := &lint.Config{
		Checks: []lint.Check{
			func(source string, spec *load.SpecInfo) []*lint.Error {
				return []*lint.Error{
					{Id: "z", Level: lint.LEVEL_ERROR, Source: "a"},
					{Id: "a", Level: lint.LEVEL_ERROR, Source: "a"},
				}
			},
		},
	}
	const source = "../data/lint/openapi.yaml"
	result := lint.Run(config, source, loadFrom(t, source))
	require.Len(t, result, 2)
	require.Equal(t, "a", result[0].Id)
	require.Equal(t, "z", result[1].Id)
}

func TestRun_SortByText(t *testing.T) {
	config := &lint.Config{
		Checks: []lint.Check{
			func(source string, spec *load.SpecInfo) []*lint.Error {
				return []*lint.Error{
					{Id: "a", Level: lint.LEVEL_ERROR, Source: "a", Text: "z"},
					{Id: "a", Level: lint.LEVEL_ERROR, Source: "a", Text: "a"},
				}
			},
		},
	}
	const source = "../data/lint/openapi.yaml"
	result := lint.Run(config, source, loadFrom(t, source))
	require.Len(t, result, 2)
	require.Equal(t, "a", result[0].Text)
	require.Equal(t, "z", result[1].Text)
}

func TestRun_SortByComment(t *testing.T) {
	config := &lint.Config{
		Checks: []lint.Check{
			func(source string, spec *load.SpecInfo) []*lint.Error {
				return []*lint.Error{
					{Id: "a", Level: lint.LEVEL_ERROR, Source: "a", Text: "a", Comment: "z"},
					{Id: "a", Level: lint.LEVEL_ERROR, Source: "a", Text: "a", Comment: "a"},
				}
			},
		},
	}
	const source = "../data/lint/openapi.yaml"
	result := lint.Run(config, source, loadFrom(t, source))
	require.Len(t, result, 2)
	require.Equal(t, "a", result[0].Comment)
	require.Equal(t, "z", result[1].Comment)
}
