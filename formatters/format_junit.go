package formatters

import (
	"encoding/xml"
	"fmt"

	"github.com/oasdiff/oasdiff/checker"
)

type JUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	TestSuites []JUnitTestSuite `xml:"testsuites"`
}

type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Package   string          `xml:"package,attr"`
	Time      string          `xml:"time,attr"`
	Tests     int             `xml:"tests,attr"`
	Errors    int             `xml:"errors,attr"`
	Failures  int             `xml:"failures,attr"`
	Name      string          `xml:"name,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	CDATA   string `xml:",innerxml"`
}

type JUnitFormatter struct {
	notImplementedFormatter
	Localizer checker.Localizer
}

func newJUnitFormatter(l checker.Localizer) JUnitFormatter {
	return JUnitFormatter{
		Localizer: l,
	}
}

func (f JUnitFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, _, _ string) ([]byte, error) {
	var testSuite = JUnitTestSuite{
		Package:   "com.oasdiff",
		Time:      "0",
		Tests:     len(changes), // TODO: use GetAllRules for the test count / test case list in the future, once the list is complete
		Errors:    0,
		Failures:  len(changes),
		Name:      "OASDiff",
		TestCases: []JUnitTestCase{},
	}

	for _, change := range changes {
		testCase := JUnitTestCase{
			Name:      change.GetId(),
			Classname: "OASDiff",
			Time:      "0",
			Failure: &JUnitFailure{
				Message: "Breaking change detected",
				CDATA:   change.GetUncolorizedText(f.Localizer),
			},
		}
		testSuite.TestCases = append(testSuite.TestCases, testCase)
	}

	// if there are no changes, add a dummy test case to the test suite as we need at least one test case
	// TODO: remove once GetAllRules is used for the test case list
	if len(changes) == 0 {
		testCase := JUnitTestCase{
			Name:      "no breaking changes detected",
			Classname: "OASDiff",
			Time:      "0",
		}
		testSuite.TestCases = append(testSuite.TestCases, testCase)
	}

	testSuites := JUnitTestSuites{TestSuites: []JUnitTestSuite{testSuite}}
	output, err := xml.MarshalIndent(testSuites, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal junit XML: %w", err)
	}

	return []byte(xml.Header + string(output)), nil
}

func (f JUnitFormatter) SupportedOutputs() []Output {
	return []Output{OutputChangelog}
}
