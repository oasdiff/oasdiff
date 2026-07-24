package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// responseHeaderSchema returns the X-RateLimit-Limit response header schema on
// the install-command GET's "default" response (openapi-test1.yaml), which the
// response-header schema checks walk.
func responseHeaderSchema(s *load.SpecInfo) *openapi3.Schema {
	return s.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Schema.Value
}

// The motivating case (#1094): a response header changing from a string with a
// parsed format to an integer (Retry-After string/date-time -> integer) drops a
// format a client parses, so it is breaking and reported at ERR.
func TestResponseHeaderTypeChanged_StringDateTimeToInteger(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	base := responseHeaderSchema(s1)
	base.Type = &openapi3.Types{"string"}
	base.Format = "date-time"
	rev := responseHeaderSchema(s2)
	rev.Type = &openapi3.Types{"integer"}
	rev.Format = ""

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseHeaderTypeChangedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t,
		"the `X-RateLimit-Limit` response header's `type/format` changed from `string/date-time` to `integer` for the status `default`",
		errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// Widening a response header's type (integer -> number, the server may now
// return values the client can't handle) is breaking, reported as generalized.
func TestResponseHeaderTypeChanged_IntegerToNumberIsGeneralized(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	responseHeaderSchema(s1).Type = &openapi3.Types{"integer"}
	responseHeaderSchema(s2).Type = &openapi3.Types{"number"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseHeaderTypeGeneralizedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// Narrowing a response header's type (number -> integer, the server returns a
// subset) is safe, reported as specialized at INFO.
func TestResponseHeaderTypeChanged_NumberToIntegerIsSpecialized(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	responseHeaderSchema(s1).Type = &openapi3.Types{"number"}
	responseHeaderSchema(s2).Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, checker.ResponseHeaderTypeSpecializedId)
}

// A bare string -> integer with no format is not breaking: a header value is
// text on the wire, so "123" is still valid for a client reading a string. It is
// a loosely typed compatible change, reported at INFO.
func TestResponseHeaderTypeChanged_StringToIntegerIsCompatible(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	responseHeaderSchema(s1).Type = &openapi3.Types{"string"}
	responseHeaderSchema(s2).Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Empty(t, checker.CheckBackwardCompatibility(allChecksConfig(), d, osm))
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseHeaderTypeCompatibleId, errs[0].GetId())
	// The header-specific comment (not the shared media-type/XML one): a header
	// value is text on the wire, so the type swap is safe.
	require.Contains(t, errs[0].GetComment(checker.NewDefaultLocalizer()), "transmitted as text")
}
