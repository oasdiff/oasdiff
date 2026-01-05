package diff_test

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"go.yaml.in/yaml/v3"
)

func ExampleGet() {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := loader.LoadFromFile("../data/simple1.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec: %v", err)
		return
	}

	s2, err := loader.LoadFromFile("../data/simple2.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load spec: %v", err)
		return
	}

	diffReport, err := diff.Get(diff.NewConfig(), s1, s2)

	if err != nil {
		fmt.Fprintf(os.Stderr, "diff failed with %v", err)
		return
	}

	bytes, err := yaml.Marshal(diffReport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal result with %v", err)
		return
	}
	fmt.Printf("%s\n", bytes)

	// Output:
	// paths:
	//     modified:
	//         /api/test:
	//             operations:
	//                 added:
	//                     - POST
	//                 deleted:
	//                     - GET
	// endpoints:
	//     added:
	//         - method: POST
	//           path: /api/test
	//     deleted:
	//         - method: GET
	//           path: /api/test
}
