package validate

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

const requestOnlyReadOnlySpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /items:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'
      responses: { "200": { description: ok } }
components:
  schemas:
    Item:
      type: object
      properties:
        id:
          type: string
          readOnly: true
`

func TestLint_ReadOnlyOnlyInRequests(t *testing.T) {
	findings := lintMutabilityReachability(mustLoad(t, requestOnlyReadOnlySpec), "spec.yaml")
	require.Len(t, findings, 1)
	require.Equal(t, ReadOnlyOnlyInRequestsID, findings[0].Id)
	require.Equal(t, checker.WARN, findings[0].Level)
	require.Contains(t, findings[0].Text, `"id"`)
}

// The same component referenced by a response as well clears the finding:
// the readOnly property has a side where it appears.
func TestLint_ReadOnlySharedWithResponse(t *testing.T) {
	spec := mustLoad(t, `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /items:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Item'
components:
  schemas:
    Item:
      type: object
      properties:
        id:
          type: string
          readOnly: true
`)
	require.Empty(t, lintMutabilityReachability(spec, "spec.yaml"))
}

// The mirror: a writeOnly property in a schema only responses use.
func TestLint_WriteOnlyOnlyInResponses(t *testing.T) {
	spec := mustLoad(t, `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /items:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  secret:
                    type: string
                    writeOnly: true
`)
	findings := lintMutabilityReachability(spec, "spec.yaml")
	require.Len(t, findings, 1)
	require.Equal(t, WriteOnlyOnlyInResponsesID, findings[0].Id)
}

// A readOnly property in a response-only schema is the flag used as
// intended; nothing to report. An orphan component is not reported either:
// reachable from nowhere is not reachable only from requests.
func TestLint_ReadOnlyUsedAsIntended(t *testing.T) {
	spec := mustLoad(t, `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /items:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Item'
components:
  schemas:
    Item:
      type: object
      properties:
        id:
          type: string
          readOnly: true
    Orphan:
      type: object
      properties:
        stale:
          type: string
          readOnly: true
`)
	require.Empty(t, lintMutabilityReachability(spec, "spec.yaml"))
}

// Reachability descends into sub-schemas: a readOnly property nested under
// an allOf branch of a request-only schema is found.
func TestLint_ReadOnlyNestedUnderAllOf(t *testing.T) {
	spec := mustLoad(t, `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /items:
    post:
      requestBody:
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    id:
                      type: string
                      readOnly: true
      responses: { "200": { description: ok } }
`)
	findings := lintMutabilityReachability(spec, "spec.yaml")
	require.Len(t, findings, 1)
	require.Equal(t, ReadOnlyOnlyInRequestsID, findings[0].Id)
}
