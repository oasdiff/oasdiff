package lint

import (
	"cmp"
	"slices"

	"github.com/oasdiff/oasdiff/load"
)

const (
	LEVEL_ERROR = 0
	LEVEL_WARN  = 1
)

type Check func(string, *load.SpecInfo) []*Error

type Error struct {
	Id      string `json:"id,omitempty" yaml:"id,omitempty"`
	Text    string `json:"text,omitempty" yaml:"text,omitempty"`
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
	Level   int    `json:"level" yaml:"level"`
	Source  string `json:"source,omitempty" yaml:"source,omitempty"`
}

type Errors []*Error

func Run(config *Config, source string, spec *load.SpecInfo) Errors {
	result := make(Errors, 0)

	if spec == nil {
		return result
	}

	for _, check := range config.Checks {
		errs := check(source, spec)
		result = append(result, errs...)
	}

	slices.SortFunc(result, func(a, b *Error) int {
		if c := cmp.Compare(a.Level, b.Level); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Source, b.Source); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Id, b.Id); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Text, b.Text); c != 0 {
			return c
		}
		return cmp.Compare(a.Comment, b.Comment)
	})
	return result
}
