package formatters

import "github.com/oasdiff/oasdiff/checker"

type ChangeGroup struct {
	Path      string
	Operation string
}

type ChangesByGroup map[ChangeGroup]*Changes

func GroupChanges(changes checker.Changes, l checker.Localizer) ChangesByGroup {

	result := ChangesByGroup{}

	for _, change := range changes {
		var group ChangeGroup

		switch change.(type) {
		case checker.ApiChange:
			group = ChangeGroup{Path: change.GetPath(), Operation: change.GetOperation()}
		default:
			group = ChangeGroup{Path: change.GetSection()}
		}

		changeEntry := Change{
			IsBreaking: change.IsBreaking(),
			Text:       change.GetUncolorizedText(l),
			Comment:    change.GetComment(l),
		}

		if c, ok := result[group]; ok {
			*c = append(*c, changeEntry)
		} else {
			result[group] = &Changes{changeEntry}
		}
	}

	return result
}
