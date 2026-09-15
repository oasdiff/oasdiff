package formatters

// Explanation is one check explained: what it reports, its severity with the
// derivation that produced it, and where in the OpenAPI document it applies.
// Validate rules carry only an id, level, and description; the other fields
// are omitted rather than rendered as empty strings.
type Explanation struct {
	Id          string   `json:"id" yaml:"id"`
	Level       string   `json:"level" yaml:"level"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Derived     bool     `json:"derived" yaml:"derived"`
	Reasoning   []string `json:"reasoning,omitempty" yaml:"reasoning,omitempty"`
	Direction   string   `json:"direction,omitempty" yaml:"direction,omitempty"`
	Area        string   `json:"area,omitempty" yaml:"area,omitempty"`
	Kind        string   `json:"kind,omitempty" yaml:"kind,omitempty"`
	Effect      string   `json:"effect,omitempty" yaml:"effect,omitempty"`
	Guards      []string `json:"guards,omitempty" yaml:"guards,omitempty"`
	Generated   bool     `json:"generated,omitempty" yaml:"generated,omitempty"`
	Locations   []string `json:"locations,omitempty" yaml:"locations,omitempty"`
	Mitigation  string   `json:"mitigation,omitempty" yaml:"mitigation,omitempty"`
	Override    string   `json:"override,omitempty" yaml:"override,omitempty"`
}
