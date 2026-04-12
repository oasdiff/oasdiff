package report_test

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/report"
)

func ExampleGetTextReportAsString() {
	swaggerLoader := openapi3.NewLoader()
	swaggerLoader.IsExternalRefsAllowed = true

	s1, err := swaggerLoader.LoadFromFile("../data/openapi-test1.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec with %v", err)
		return
	}

	s2, err := swaggerLoader.LoadFromFile("../data/openapi-test3.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec with %v", err)
		return
	}

	diffReport, err := diff.Get(&diff.Config{}, s1, s2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff failed with %v", err)
		return
	}

	fmt.Print(report.GetTextReportAsString(diffReport))

	// Output:
	// ### New Endpoints: None
	// -----------------------
	//
	// ### Deleted Endpoints: None
	// ---------------------------
	//
	// ### Modified Endpoints: 4
	// -------------------------
	// GET /api/{domain}/{project}/badges/security-score
	// - Extensions changed
	//   - Deleted extension: x-extension-test
	// - Tags changed from 'security' to ''
	// - OperationID changed from 'GetSecurityScores' to 'GetSecurityScore'
	// - Deleted query param: filter
	// - Deleted header param: user
	// - Deleted cookie param: test
	// - Modified query param: image
	//   - Extensions changed
	//     - New extension: x-extension-test
	//   - Schema changed
	//     - Extensions changed
	//       - Deleted extension: x-extension-test
	//     - Property 'Not' changed
	//       - Schema added
	//     - Type changed from 'string' to ''
	//     - Format changed from 'general string' to ''
	//     - Description changed from 'alphanumeric' to ''
	//     - Example changed from 'tufinim/generic-bank:cia-latest' to null
	//     - Pattern changed from '^(?:[\w-./:]+)$' to ''
	//   - Examples changed
	//     - New example: 1
	//     - Modified example: 0
	//       - Value changed from 'reuven' to 'reuven1'
	// - Modified query param: token
	//   - Schema changed
	//     - Example changed from '26734565-dbcc-449a-a370-0beaaf04b0e8' to '26734565-dbcc-449a-a370-0beaaf04b0e7'
	//     - MaxLength changed from 29 to 30
	// - Responses changed
	//   - New response: default
	//   - Deleted response: 200
	//   - Deleted response: 201
	//   - Deleted response: 400
	//
	// GET /api/{domain}/{project}/install-command
	// - Deleted header param: network-policies
	// - Modified path param: project
	//   - Schema changed
	//     - New enum values: [test1]
	// - Responses changed
	//   - Modified response: default
	//     - Extensions changed
	//       - New extension: x-extension-test
	//       - New extension: x-test
	//     - Description changed from 'Tufin1' to 'Tufin'
	//     - Headers changed
	//       - Modified header: X-RateLimit-Limit
	//         - Extensions changed
	//           - New extension: x-test
	//         - Description changed from 'Request limit per hour.' to 'Request limit per min.'
	// - Servers changed
	//   - New server: https://api.oasdiff.com
	//   - New server: https://www.oasdiff.com
	//
	// POST /register
	// - Callbacks changed
	// - Security changed
	//   - Deleted security requirements: bearerAuth
	//   - Modified security requirements: OAuth
	//     - Scheme OAuth Added scopes: [write:pets]
	//
	// POST /subscribe
	// - Callbacks changed
	//
	// Other Changes
	// -------------
	// Extensions changed
	// - Deleted extension: x-extension-test
	// - Modified extension: x-extension-test2
	//   - Modified value from 'go' to 'nogo'
	//
	// Security Requirements changed
	// - Deleted security requirements: bearerAuth
	//
	// Servers changed
	// - Deleted server: tufin.com
	//
}

func ExampleGetHTMLReportAsString() {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := loader.LoadFromFile("../data/openapi-test1.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec with %v", err)
		return
	}

	s2, err := loader.LoadFromFile("../data/openapi-test3.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec with %v", err)
		return
	}

	diffReport, err := diff.Get(&diff.Config{}, s1, s2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff failed with %v", err)
		return
	}

	html, err := report.GetHTMLReportAsString(diffReport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate HTML with %v", err)
		return
	}

	fmt.Print(html)

	// Output:
	// <h3 id="new-endpoints-none">New Endpoints: None</h3>
	// <hr>
	// <h3 id="deleted-endpoints-none">Deleted Endpoints: None</h3>
	// <hr>
	// <h3 id="modified-endpoints-4">Modified Endpoints: 4</h3>
	// <hr>
	// <p>GET /api/{domain}/{project}/badges/security-score</p>
	// <ul>
	// <li>Extensions changed
	// <ul>
	// <li>Deleted extension: x-extension-test</li>
	// </ul>
	// </li>
	// <li>Tags changed from 'security' to ''</li>
	// <li>OperationID changed from 'GetSecurityScores' to 'GetSecurityScore'</li>
	// <li>Deleted query param: filter</li>
	// <li>Deleted header param: user</li>
	// <li>Deleted cookie param: test</li>
	// <li>Modified query param: image
	// <ul>
	// <li>Extensions changed
	// <ul>
	// <li>New extension: x-extension-test</li>
	// </ul>
	// </li>
	// <li>Schema changed
	// <ul>
	// <li>Extensions changed
	// <ul>
	// <li>Deleted extension: x-extension-test</li>
	// </ul>
	// </li>
	// <li>Property 'Not' changed
	// <ul>
	// <li>Schema added</li>
	// </ul>
	// </li>
	// <li>Type changed from 'string' to ''</li>
	// <li>Format changed from 'general string' to ''</li>
	// <li>Description changed from 'alphanumeric' to ''</li>
	// <li>Example changed from 'tufinim/generic-bank:cia-latest' to null</li>
	// <li>Pattern changed from '^(?:[\w-./:]+)$' to ''</li>
	// </ul>
	// </li>
	// <li>Examples changed
	// <ul>
	// <li>New example: 1</li>
	// <li>Modified example: 0
	// <ul>
	// <li>Value changed from 'reuven' to 'reuven1'</li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// <li>Modified query param: token
	// <ul>
	// <li>Schema changed
	// <ul>
	// <li>Example changed from '26734565-dbcc-449a-a370-0beaaf04b0e8' to '26734565-dbcc-449a-a370-0beaaf04b0e7'</li>
	// <li>MaxLength changed from 29 to 30</li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// <li>Responses changed
	// <ul>
	// <li>New response: default</li>
	// <li>Deleted response: 200</li>
	// <li>Deleted response: 201</li>
	// <li>Deleted response: 400</li>
	// </ul>
	// </li>
	// </ul>
	// <p>GET /api/{domain}/{project}/install-command</p>
	// <ul>
	// <li>Deleted header param: network-policies</li>
	// <li>Modified path param: project
	// <ul>
	// <li>Schema changed
	// <ul>
	// <li>New enum values: [test1]</li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// <li>Responses changed
	// <ul>
	// <li>Modified response: default
	// <ul>
	// <li>Extensions changed
	// <ul>
	// <li>New extension: x-extension-test</li>
	// <li>New extension: x-test</li>
	// </ul>
	// </li>
	// <li>Description changed from 'Tufin1' to 'Tufin'</li>
	// <li>Headers changed
	// <ul>
	// <li>Modified header: X-RateLimit-Limit
	// <ul>
	// <li>Extensions changed
	// <ul>
	// <li>New extension: x-test</li>
	// </ul>
	// </li>
	// <li>Description changed from 'Request limit per hour.' to 'Request limit per min.'</li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// <li>Servers changed
	// <ul>
	// <li>New server: <a href="https://api.oasdiff.com">https://api.oasdiff.com</a></li>
	// <li>New server: <a href="https://www.oasdiff.com">https://www.oasdiff.com</a></li>
	// </ul>
	// </li>
	// </ul>
	// <p>POST /register</p>
	// <ul>
	// <li>Callbacks changed</li>
	// <li>Security changed
	// <ul>
	// <li>Deleted security requirements: bearerAuth</li>
	// <li>Modified security requirements: OAuth
	// <ul>
	// <li>Scheme OAuth Added scopes: [write:pets]</li>
	// </ul>
	// </li>
	// </ul>
	// </li>
	// </ul>
	// <p>POST /subscribe</p>
	// <ul>
	// <li>Callbacks changed</li>
	// </ul>
	// <h2 id="other-changes">Other Changes</h2>
	// <p>Extensions changed</p>
	// <ul>
	// <li>Deleted extension: x-extension-test</li>
	// <li>Modified extension: x-extension-test2
	// <ul>
	// <li>Modified value from 'go' to 'nogo'</li>
	// </ul>
	// </li>
	// </ul>
	// <p>Security Requirements changed</p>
	// <ul>
	// <li>Deleted security requirements: bearerAuth</li>
	// </ul>
	// <p>Servers changed</p>
	// <ul>
	// <li>Deleted server: tufin.com</li>
	// </ul>
}
