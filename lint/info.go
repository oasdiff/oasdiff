package lint

import (
	"fmt"
	"net/url"

	"github.com/tufin/oasdiff/load"
)

func InfoCheck(source string, spec *load.OpenAPISpecInfo) []*Error {

	result := make([]*Error, 0)

	if spec == nil || spec.Spec == nil {
		return result
	}

	if spec.Spec.Info == nil {
		result = append(result, &Error{
			Id:      "info-missing",
			Level:   LEVEL_ERROR,
			Text:    "Info is missing",
			Comment: "It is a good practice to include general information about your API into the specification. Title and Version fields are required.",
			Source:  source,
		})
		return result
	}

	if spec.Spec.Info.Title == "" {
		result = append(result, &Error{
			Id:     "info-title-missing",
			Level:  LEVEL_ERROR,
			Text:   "The title of the API is missing",
			Source: source,
		})
	}
	if spec.Spec.Info.Version == "" {
		result = append(result, &Error{
			Id:     "info-version-missing",
			Level:  LEVEL_ERROR,
			Text:   "The version of the API is missing",
			Source: source,
		})
	}

	if tos := spec.Spec.Info.TermsOfService; tos != "" {
		if _, err := url.ParseRequestURI(tos); err != nil {
			result = append(result, &Error{
				Id:     "info-invalud-terms-of-service",
				Level:  LEVEL_ERROR,
				Text:   fmt.Sprintf("Terms of service must be in the format of a URL: %s", tos),
				Source: source,
			})
		}
	}

	return result
}
