package internal

// documentView holds the options of diff and summary that change the specs
// or leave parts of them out of the comparison. The checks compare the specs
// in full and as written, so breaking and changelog use the zero value.
type documentView struct {
	lowercaseHeaders  bool
	flattenParams     bool
	excludeElements   []string
	excludeExtensions []string
}

func (flags *Flags) getDocumentView() documentView {
	return documentView{
		lowercaseHeaders:  flags.getCaseInsensitiveHeaders(),
		flattenParams:     flags.getFlattenParams(),
		excludeElements:   flags.getExcludeElements(),
		excludeExtensions: flags.getExcludeExtensions(),
	}
}
