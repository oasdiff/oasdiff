package metaschema

import "strings"

// Polarity is the syntactic position of a location in the document.
type Polarity string

const (
	PolarityRequest  Polarity = "request"
	PolarityResponse Polarity = "response"
	PolarityShared   Polarity = "shared"   // components: request or response depending on the referencing site
	PolarityDocument Polarity = "document" // neither wire direction
)

func polarity(path string) Polarity {
	p := PolarityDocument
	if strings.HasPrefix(path, "components.") {
		p = PolarityShared
	}
	for seg := range strings.SplitSeq(path, ".") {
		switch seg {
		case "parameters", "requestBody":
			p = PolarityRequest
		case "responses":
			p = PolarityResponse
		}
	}
	return p
}
