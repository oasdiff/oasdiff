package checker_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

func ExampleCheckBackwardCompatibility() {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/openapi-test1.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec: %v", err)
		return
	}

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/openapi-test3.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec: %v", err)
		return
	}

	diffRes, operationsSources, err := diff.GetPathsDiff(diff.NewConfig(),
		[]*load.SpecInfo{s1},
		[]*load.SpecInfo{s2},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "diff failed with %v", err)
		return
	}

	errs := checker.CheckBackwardCompatibility(checker.NewConfig(checker.GetAllChecks()), diffRes, operationsSources)

	// process configuration file for ignoring errors
	errs, err = checker.ProcessIgnoredBackwardCompatibilityErrors(checker.ERR, errs, "../data/ignore-err-example.txt", checker.NewDefaultLocalizer())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ignore errors failed with %v", err)
		return
	}

	// process configuration file for ignoring warnings
	errs, err = checker.ProcessIgnoredBackwardCompatibilityErrors(checker.WARN, errs, "../data/ignore-warn-example.txt", checker.NewDefaultLocalizer())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ignore warnings failed with %v", err)
		return
	}

	// pretty print breaking changes errors
	if len(errs) > 0 {
		localizer := checker.NewDefaultLocalizer()
		count := errs.GetLevelCount()
		fmt.Print(localizer("total-errors", len(errs), count[checker.ERR], "error", count[checker.WARN], "warning"))
		for _, bcerr := range errs {
			fmt.Printf("%s\n\n", strings.TrimRight(bcerr.SingleLineError(localizer, checker.ColorNever), " "))
		}
	}

	// Output:
	// 4 breaking changes: 1 error, 3 warning
	// error at ../data/openapi-test3.yaml, in API GET /api/{domain}/{project}/badges/security-score removed the success response with the status '201' [response-success-status-removed].
	//
	// warning at ../data/openapi-test3.yaml, in API GET /api/{domain}/{project}/badges/security-score deleted the 'cookie' request parameter 'test' [request-parameter-removed]. This is a warning because some apps may return an error when receiving a parameter that they do not expect. It is recommended to deprecate the parameter first.
	//
	// warning at ../data/openapi-test3.yaml, in API GET /api/{domain}/{project}/badges/security-score deleted the 'header' request parameter 'user' [request-parameter-removed]. This is a warning because some apps may return an error when receiving a parameter that they do not expect. It is recommended to deprecate the parameter first.
	//
	// warning at ../data/openapi-test3.yaml, in API GET /api/{domain}/{project}/badges/security-score deleted the 'query' request parameter 'filter' [request-parameter-removed]. This is a warning because some apps may return an error when receiving a parameter that they do not expect. It is recommended to deprecate the parameter first.
}
