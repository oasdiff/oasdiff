package diff

type MediaTypeNameDiff struct {
	NameDiff       *ValueDiff     `json:"name,omitempty" yaml:"name,omitempty"`
	TypeDiff       *ValueDiff     `json:"type,omitempty" yaml:"type,omitempty"`
	SubtypeDiff    *ValueDiff     `json:"subtype,omitempty" yaml:"subtype,omitempty"`
	SuffixDiff     *ValueDiff     `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	ParametersDiff *StringMapDiff `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	contained      bool
}

func (diff *MediaTypeNameDiff) Empty() bool {
	return diff == nil || *diff == MediaTypeNameDiff{}
}

func getMediaTypeNameDiff(name1, name2 string) (*MediaTypeNameDiff, error) {
	diff, err := getMediaTypeNameDiffInternal(name1, name2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getMediaTypeNameDiffInternal(name1, name2 string) (*MediaTypeNameDiff, error) {
	if name1 == name2 {
		return nil, nil
	}

	mediaTypeName1, err := ParseMediaTypeName(name1)
	if err != nil {
		return nil, err
	}

	mediaTypeName2, err := ParseMediaTypeName(name2)
	if err != nil {
		return nil, err
	}

	return &MediaTypeNameDiff{
		NameDiff:       getValueDiff(name1, name2),
		TypeDiff:       getValueDiff(mediaTypeName1.Type, mediaTypeName2.Type),
		SubtypeDiff:    getValueDiff(mediaTypeName1.Subtype, mediaTypeName2.Subtype),
		SuffixDiff:     getValueDiff(mediaTypeName1.Suffix, mediaTypeName2.Suffix),
		ParametersDiff: getStringMapDiff(mediaTypeName1.Parameters, mediaTypeName2.Parameters),
		contained:      IsMediaTypeNameContained(name1, name2),
	}, nil
}

func (diff *MediaTypeNameDiff) IsContained() bool {
	return diff.contained
}
