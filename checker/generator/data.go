package generator

import "slices"

func getAll() ValueSets {
	return slices.Concat(
		getRequest(),
		getResponse(),
	)
}

func getRequest() ValueSets {
	return slices.Concat(
		getSchema([]string{"media-type", "request body"}, nil),
		getSchema([]string{"property", "media-type", "request body"}, nil),
		getSchema([]string{"request parameter"}, []bool{true}),
	)
}

func getResponse() ValueSets {
	return slices.Concat(
		getSchema([]string{"media-type", "response"}, nil),
		getSchema([]string{"property", "media-type", "response"}, nil),
	)
}

func getSchema(hierarchy []string, attributed []bool) ValueSets {
	return ValueSets{
		ValueSetA{
			adjective:  "value",
			hierarchy:  hierarchy,
			attributed: attributed,
			nouns:      []string{"max", "maxLength", "min", "minLength", "minItems", "maxItems"},
			actions:    []string{"set", "increase", "decrease"},
		},
		ValueSetA{
			hierarchy: hierarchy,
			nouns:     []string{"type/format"},
			actions:   []string{"change", "generalize"},
		},
		ValueSetA{
			hierarchy: hierarchy,
			nouns:     []string{"discriminator property name"},
			actions:   []string{"change"},
		},
		ValueSetB{
			hierarchy: append([]string{"anyOf list"}, hierarchy...),
			nouns:     []string{"%s"},
			actions:   []string{"add", "remove"},
		},
		ValueSetB{
			hierarchy: append([]string{"oneOf list"}, hierarchy...),
			nouns:     []string{"%s"},
			actions:   []string{"add", "remove"},
		},
		ValueSetB{
			hierarchy: append([]string{"allOf list"}, hierarchy...),
			nouns:     []string{"%s"},
			actions:   []string{"add", "remove"},
		},
		ValueSetB{
			hierarchy: hierarchy,
			nouns:     []string{"discriminator", "mapping keys"},
			actions:   []string{"add", "remove"},
		},
	}
}
