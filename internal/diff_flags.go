package internal

import (
	"github.com/tufin/oasdiff/diff"
)

type DiffFlags struct {
	base                     string
	revision                 string
	composed                 bool
	prefixBase               string
	prefixRevision           string
	stripPrefixBase          string
	stripPrefixRevision      string
	matchPath                string
	filterExtension          string
	format                   string
	failOnDiff               bool
	circularReferenceCounter int
	matchPathParams          bool
	excludeElements          []string
}

func (flags *DiffFlags) toConfig() *diff.Config {
	config := diff.NewConfig()
	config.PathFilter = flags.matchPath
	config.FilterExtension = flags.filterExtension
	config.PathPrefixBase = flags.prefixBase
	config.PathPrefixRevision = flags.prefixRevision
	config.PathStripPrefixBase = flags.stripPrefixBase
	config.PathStripPrefixRevision = flags.stripPrefixRevision
	config.MatchPathParams = flags.matchPathParams
	config.SetExcludeElements(flags.excludeElements)

	return config
}

// func (flags *DiffFlags) validate() *ReturnError {
// 	if flags.format == "json" && !flags.getExcludeEndpoints() {
// 		return getErrInvalidFlags(fmt.Errorf("json format requires \"-exclude-elements endpoints\""))
// 	}

// 	return nil
// }
