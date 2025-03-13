package formatters

import "github.com/oasdiff/oasdiff/checker"

type Endpoint struct {
	Path      string
	Operation string
}

type ChangesByEndpoint map[Endpoint]*Changes

func GroupChanges(changes checker.Changes, l checker.Localizer) ChangesByEndpoint {

	apiChanges := ChangesByEndpoint{}

	for _, change := range changes {
		switch change.(type) {
		case checker.ApiChange:
			ep := Endpoint{Path: change.GetPath(), Operation: change.GetOperation()}
			if c, ok := apiChanges[ep]; ok {
				*c = append(*c, Change{
					IsBreaking: change.IsBreaking(),
					Text:       change.GetUncolorizedText(l),
					Level:      change.GetLevel(),
				})
			} else {
				apiChanges[ep] = &Changes{Change{
					IsBreaking: change.IsBreaking(),
					Text:       change.GetUncolorizedText(l),
					Level:      change.GetLevel(),
				}}
			}
		}
	}

	return apiChanges
}
