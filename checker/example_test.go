package checker_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
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

	// render the breaking changes, which is the formatters package's job
	if len(errs) > 0 {
		formatter, err := formatters.Lookup(string(formatters.FormatSingleLine), formatters.DefaultFormatterOpts())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to look up the formatter: %v", err)
			return
		}

		opts := formatters.NewRenderOpts()
		opts.ColorMode = checker.ColorNever
		report, err := formatter.RenderChangelog(errs, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to render the report: %v", err)
			return
		}
		// a change with no comment ends in a space, which this example's
		// expected output cannot carry
		fmt.Print(strings.ReplaceAll(string(report), " \n", "\n"))
	}

	// Output:
	// 6 changes: 3 error, 3 warning, 0 info
	// error at ../data/openapi-test1.yaml, in API GET /api/{domain}/{project}/badges/security-score removed the success response with the status `201` [response-success-status-removed].
	//
	// error at ../data/openapi-test1.yaml, in API POST /register the endpoint scheme security `bearerAuth` was removed from the API [api-security-removed].
	//
	// error at ../data/openapi-test3.yaml, in API POST /register the security scope `write:pets` was added to the endpoint's security scheme `OAuth` [api-security-scope-added].
	//
	// warning at ../data/openapi-test1.yaml, in API GET /api/{domain}/{project}/badges/security-score deleted the `cookie` request parameter `test` [request-parameter-removed]. This is a warning because some clients may return an error when receiving an unexpected parameter. It is recommended to deprecate the parameter first.
	//
	// warning at ../data/openapi-test1.yaml, in API GET /api/{domain}/{project}/badges/security-score deleted the `header` request parameter `user` [request-parameter-removed]. This is a warning because some clients may return an error when receiving an unexpected parameter. It is recommended to deprecate the parameter first.
	//
	// warning at ../data/openapi-test1.yaml, in API GET /api/{domain}/{project}/badges/security-score deleted the `query` request parameter `filter` [request-parameter-removed]. This is a warning because some clients may return an error when receiving an unexpected parameter. It is recommended to deprecate the parameter first.
}
