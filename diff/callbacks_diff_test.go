package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func loadCallbackCycle(t *testing.T, file string) *openapi3.T {
	t.Helper()
	spec, err := openapi3.NewLoader().LoadFromFile("../data/callback_cycle/" + file)
	require.NoError(t, err)
	return spec
}

// The callback's operation carries the same callback, so comparing it reaches
// the pair being compared again. Its changes are reported once, where the
// callback is first compared.
func TestDiff_CallbackCycle(t *testing.T) {
	d, err := diff.Get(diff.NewConfig(), loadCallbackCycle(t, "base.yaml"), loadCallbackCycle(t, "revision.yaml"))
	require.NoError(t, err)

	operation := d.PathsDiff.Modified["/x"].OperationsDiff.Modified["POST"]
	callbackOperation := operation.CallbacksDiff.Modified["cb"].Modified["{$request.body#/u}"].OperationsDiff.Modified["POST"]
	require.Equal(t, &diff.ValueDiff{From: "inner before", To: "inner after"}, callbackOperation.DescriptionDiff)
	require.Nil(t, callbackOperation.CallbacksDiff)
}

func TestDiff_CallbackCycleSame(t *testing.T) {
	d, err := diff.Get(diff.NewConfig(), loadCallbackCycle(t, "base.yaml"), loadCallbackCycle(t, "base.yaml"))
	require.NoError(t, err)
	require.Nil(t, d)
}
