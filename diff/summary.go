package diff

import "reflect"

// Summary summarizes the changes between a pair of OpenAPI specifications
type Summary struct {
	Diff    bool                           `json:"diff" yaml:"diff"`
	Details map[DetailName]*SummaryDetails `json:"details,omitempty" yaml:"details,omitempty"`
}

func newSummary() *Summary {
	return &Summary{
		Details: map[DetailName]*SummaryDetails{},
	}
}

// SummaryDetails summarizes the changes between equivalent parts of the two OpenAPI specifications: paths, schemas, parameters, headers, responses etc.
type SummaryDetails struct {
	Added    int `json:"added,omitempty" yaml:"added,omitempty"`       // number of added items
	Deleted  int `json:"deleted,omitempty" yaml:"deleted,omitempty"`   // number of deleted items
	Modified int `json:"modified,omitempty" yaml:"modified,omitempty"` // number of modified items
}

type detailWithSummary interface {
	getSummary() *SummaryDetails
}

// DetailName is the key type of the summary map
type DetailName string

// Detail constants are the keys in the summary map
const (
	// Swagger
	PathsDetail        DetailName = "paths"
	SecurityDetail     DetailName = "security"
	ServersDetail      DetailName = "servers"
	TagsDetail         DetailName = "tags"
	ExternalDocsDetail DetailName = "externalDocs"

	// Components
	SchemasDetail         DetailName = "schemas"
	ParametersDetail      DetailName = "parameters"
	HeadersDetail         DetailName = "headers"
	RequestBodiesDetail   DetailName = "requestBodies"
	ResponsesDetail       DetailName = "responses"
	SecuritySchemesDetail DetailName = "securitySchemes"
	ExamplesDetail        DetailName = "examples"
	LinksDetail           DetailName = "links"
	CallbacksDetail       DetailName = "callbacks"

	// Special
	EndpointsDetail DetailName = "endpoints"
)

// GetSummaryDetails returns the summary for a specific part
func (summary *Summary) GetSummaryDetails(name DetailName) SummaryDetails {
	if details, ok := summary.Details[name]; ok {
		if details != nil {
			return *details
		}
	}

	return SummaryDetails{}
}

func (summary *Summary) add(detail detailWithSummary, detailName DetailName) {
	if !isNilPointer(detail) {
		summary.Details[detailName] = detail.getSummary()
	}
}

func isNilPointer(i any) bool {
	return reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil()
}
