package review

import (
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// An endpoint that refers to Order from its request and its response, where
// Order refers to Address, and to Address directly from a parameter.
const referencesSpec = `openapi: 3.1.0
info:
  title: t
  version: "1"
paths:
  /orders:
    post:
      operationId: createOrder
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Order"
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Order"
    get:
      operationId: listOrders
      responses:
        "200":
          description: ok
components:
  schemas:
    Order:
      type: object
      properties:
        shipTo:
          $ref: "#/components/schemas/Address"
    Address:
      type: object
      properties:
        street:
          type: string
    Unused:
      type: object
`

// changeAt is a change to POST /orders located on the given line.
func changeAt(id string, line int) checker.ApiChange {
	return checker.ApiChange{
		Id:        id,
		Operation: "POST",
		Path:      "/orders",
		CommonChange: checker.CommonChange{
			BaseSource:     &checker.Source{Line: line},
			RevisionSource: &checker.Source{Line: line},
		},
	}
}

func extractReferences(t *testing.T, changes ...checker.Change) []Block {
	t.Helper()
	doc := loadWithOrigin(t, referencesSpec)
	return Extract(changes, docs(doc), docs(doc), oneFile("", referencesSpec), oneFile("", referencesSpec))
}

// A change located in a component the endpoint refers to is listed in the
// component's block, and the endpoint's block records the line of the $ref.
func TestExtract_ReferenceFromEndpoint(t *testing.T) {
	change := changeAt("request-property-type-changed", 30)
	blocks := extractReferences(t, change)
	require.Len(t, blocks, 2)

	require.Equal(t, "components/schemas/Order", blocks[0].Key)
	require.Equal(t, []string{"request-property-type-changed"}, blocks[0].ChangeIDs)
	require.Empty(t, blocks[0].References)

	require.Equal(t, "POST /orders", blocks[1].Key)
	require.Empty(t, blocks[1].ChangeIDs)
	require.Contains(t, blocks[1].RevText, "$ref")
	require.Equal(t, []Reference{{Fingerprint: checker.Fingerprint(change), BaseLine: 13, RevLine: 13}}, blocks[1].References)
}

// A change to a response is tied to the $ref under the endpoint's responses,
// although the request refers to the same component first.
func TestExtract_ReferenceFollowsDirection(t *testing.T) {
	blocks := extractReferences(t, changeAt("response-property-type-changed", 30))
	require.Len(t, blocks, 2)
	require.Equal(t, 20, blocks[1].References[0].RevLine)
}

// A change in a component the endpoint reaches only through another component
// is tied to the $ref in the endpoint that leads there.
func TestExtract_ReferenceThroughAnotherComponent(t *testing.T) {
	blocks := extractReferences(t, changeAt("request-property-type-changed", 37))
	require.Len(t, blocks, 2)
	require.Equal(t, "components/schemas/Address", blocks[0].Key)
	require.Equal(t, "POST /orders", blocks[1].Key)
	require.Equal(t, 13, blocks[1].References[0].RevLine)
}

// A change to a parameter is tied to the parameter's $ref and a change to the
// request body to the body's, although both lead to the same component.
func TestExtract_ReferenceFollowsEndpointPart(t *testing.T) {
	const spec = `openapi: 3.1.0
info:
  title: t
  version: "1"
paths:
  /orders:
    post:
      parameters:
        - name: status
          in: query
          schema:
            $ref: "#/components/schemas/Status"
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Status"
      responses:
        "200":
          description: ok
components:
  schemas:
    Status:
      type: string
`
	doc := loadWithOrigin(t, spec)
	extract := func(id string) []Block {
		return Extract(checker.Changes{changeAt(id, 24)}, docs(doc), docs(doc), oneFile("", spec), oneFile("", spec))
	}

	blocks := extract("request-parameter-type-changed")
	require.Len(t, blocks, 2)
	require.Equal(t, 12, blocks[1].References[0].RevLine)

	blocks = extract("request-body-type-changed")
	require.Len(t, blocks, 2)
	require.Equal(t, 17, blocks[1].References[0].RevLine)
}

// Components that refer to each other in a cycle are followed once.
func TestExtract_ReferenceThroughCircularComponents(t *testing.T) {
	const spec = `openapi: 3.1.0
info:
  title: t
  version: "1"
paths:
  /nodes:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Node"
      responses:
        "200":
          description: ok
components:
  schemas:
    Node:
      type: object
      properties:
        child:
          $ref: "#/components/schemas/Leaf"
    Leaf:
      type: object
      properties:
        parent:
          $ref: "#/components/schemas/Node"
        name:
          type: string
`
	doc := loadWithOrigin(t, spec)
	change := checker.ApiChange{
		Id:        "request-property-type-changed",
		Operation: "POST",
		Path:      "/nodes",
		CommonChange: checker.CommonChange{
			BaseSource:     &checker.Source{Line: 28},
			RevisionSource: &checker.Source{Line: 28},
		},
	}
	blocks := Extract(checker.Changes{change}, docs(doc), docs(doc), oneFile("", spec), oneFile("", spec))
	require.Len(t, blocks, 2)
	require.Equal(t, "components/schemas/Leaf", blocks[0].Key)
	require.Equal(t, "POST /nodes", blocks[1].Key)
	require.Equal(t, 12, blocks[1].References[0].RevLine)
}

// A change located in its own endpoint has no reference.
func TestExtract_NoReferenceInsideTheEndpoint(t *testing.T) {
	blocks := extractReferences(t, changeAt("request-property-type-changed", 11))
	require.Len(t, blocks, 1)
	require.Equal(t, "POST /orders", blocks[0].Key)
	require.Empty(t, blocks[0].References)
}

// A change located in a component the endpoint does not refer to has none
// either, and the endpoint's block is not extracted for it.
func TestExtract_NoReferenceToUnreachedComponent(t *testing.T) {
	blocks := extractReferences(t, changeAt("request-property-type-changed", 39))
	require.Len(t, blocks, 1)
	require.Equal(t, "components/schemas/Unused", blocks[0].Key)
}

// An endpoint with its own change and a change in a component lists the
// first and references the second.
func TestExtract_ReferenceNextToOwnChange(t *testing.T) {
	own := changeAt("request-body-became-required", 9)
	inComponent := changeAt("request-property-type-changed", 30)
	blocks := extractReferences(t, own, inComponent)
	require.Len(t, blocks, 2)
	require.Equal(t, "POST /orders", blocks[0].Key)
	require.Equal(t, []string{"request-body-became-required"}, blocks[0].ChangeIDs)
	require.Equal(t, checker.Fingerprint(inComponent), blocks[0].References[0].Fingerprint)
}

// Through the real diff and checks: a oneOf whose nullability changed because
// a referenced branch became nullable is located in the branch's component,
// and the endpoint's block records the $ref inside the oneOf, the request's
// for the request change and the response's for the response change.
func TestExtract_ReferenceToOneOfBranch(t *testing.T) {
	const basePath, revPath = "../data/checker/one_of_null_branch_base.yaml", "../data/checker/one_of_null_branch_revision.yaml"
	baseText, err := os.ReadFile(basePath)
	require.NoError(t, err)
	revText, err := os.ReadFile(revPath)
	require.NoError(t, err)

	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	s1, err := load.NewSpecInfoFromData(loader, baseText, "base.yaml")
	require.NoError(t, err)
	s2, err := load.NewSpecInfoFromData(loader, revText, "revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, sources, checker.INFO)
	require.Len(t, changes, 2)

	blocks := Extract(changes, docs(s1.Spec), docs(s2.Spec), oneFile("base.yaml", string(baseText)), oneFile("revision.yaml", string(revText)))
	byKey := map[string]Block{}
	for _, b := range blocks {
		byKey[b.Key] = b
	}
	require.Len(t, byKey, 3)
	require.Len(t, byKey["components/schemas/RequestAddress"].ChangeIDs, 1)
	require.Len(t, byKey["components/schemas/ResponseAddress"].ChangeIDs, 1)

	endpoint := byKey["POST /products"]
	require.Empty(t, endpoint.ChangeIDs)
	lines := map[string]int{}
	for _, reference := range endpoint.References {
		require.Equal(t, reference.BaseLine, reference.RevLine)
		lines[reference.Fingerprint] = reference.RevLine
	}
	for _, change := range changes {
		want := 19
		if change.GetId() == checker.ResponsePropertyBecameNotNullableId {
			want = 31
		}
		require.Equal(t, want, lines[checker.Fingerprint(change)], change.GetId())
	}
}
