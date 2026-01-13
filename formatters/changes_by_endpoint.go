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
		switch apiChange := change.(type) {
		case checker.ApiChange:
			ep := Endpoint{Path: change.GetPath(), Operation: change.GetOperation()}
			// Get comment directly from ApiChange to avoid localization
			comment := ""
			if len(apiChange.Comment) > 0 && apiChange.Comment[0] == ' ' {
				comment = apiChange.Comment
			} else {
				comment = change.GetComment(l)
			}

			changeEntry := Change{
				IsBreaking: change.IsBreaking(),
				Text:       change.GetUncolorizedText(l),
				Comment:    comment,
			}

			if c, ok := apiChanges[ep]; ok {
				*c = append(*c, changeEntry)
			} else {
				apiChanges[ep] = &Changes{changeEntry}
			}
		}
	}

	return apiChanges
}
