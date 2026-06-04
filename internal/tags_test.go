package internal_test

import (
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func Test_ChecksNoTags(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru"), io.Discard, io.Discard))
}

func Test_ChecksTagsDirection(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags request"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags response"), io.Discard, io.Discard))
}

func Test_ChecksTagsAction(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags add"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags remove"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags change"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags generalize"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags specialize"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags increase"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags decrease"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags set"), io.Discard, io.Discard))
}

func Test_ChecksTagsArea(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags schema"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags parameters"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags requestBody"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags responses"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags paths"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags headers"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags security"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags tags"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags components"), io.Discard, io.Discard))
}

func Test_ChecksTagsKind(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags existence"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags requiredness"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags type"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags constraints"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags values"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags structure"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags lifecycle"), io.Discard, io.Discard))
}
