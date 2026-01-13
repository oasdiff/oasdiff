package formatters

import (
	"github.com/oasdiff/oasdiff/checker"
)

type Change struct {
	Id          string         `json:"id,omitempty" yaml:"id,omitempty"`
	Text        string         `json:"text,omitempty" yaml:"text,omitempty"`
	Comment     string         `json:"comment,omitempty" yaml:"comment,omitempty"`
	Level       checker.Level  `json:"level" yaml:"level"`
	Operation   string         `json:"operation,omitempty" yaml:"operation,omitempty"`
	OperationId string         `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Path        string         `json:"path,omitempty" yaml:"path,omitempty"`
	Source      string         `json:"source,omitempty" yaml:"source,omitempty"`
	Section     string         `json:"section,omitempty" yaml:"section,omitempty"`
	IsBreaking  bool           `json:"-" yaml:"-"`
	Attributes  map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

type Changes []Change

func NewChanges(originalChanges checker.Changes, l checker.Localizer) Changes {
	changes := make(Changes, len(originalChanges))
	for i, change := range originalChanges {
		comment := change.GetComment(l)
		// If comment looks like formatted details (starts with " ("), don't localize it
		// Just use the raw comment string from the change
		if apiChange, ok := change.(checker.ApiChange); ok {
			if len(apiChange.Comment) > 0 && apiChange.Comment[0] == ' ' {
				comment = apiChange.Comment
			}
		}
		changes[i] = Change{
			Section:     change.GetSection(),
			Id:          change.GetId(),
			Text:        change.GetUncolorizedText(l),
			Comment:     comment,
			Level:       change.GetLevel(),
			Operation:   change.GetOperation(),
			OperationId: change.GetOperationId(),
			Path:        change.GetPath(),
			Source:      change.GetSource(),
			Attributes:  change.GetAttributes(),
		}
	}
	return changes
}
