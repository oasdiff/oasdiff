package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func Test_UpgradeCmdOK(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff upgrade ../data/upgrade/simple-30.yaml"), io.Discard, io.Discard))
}

func Test_UpgradeCmdProducesCanonicalForm(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff upgrade ../data/upgrade/simple-30.yaml"), &stdout, io.Discard))
	out := stdout.String()
	// Version-string bump.
	require.Contains(t, out, "openapi: 3.2.1")
	// nullable: true -> type array including "null".
	require.Contains(t, out, "- \"null\"")
	require.NotContains(t, out, "nullable: true")
	// boolean exclusiveMinimum -> numeric form.
	require.Contains(t, out, "exclusiveMinimum: 0")
	// example -> examples.
	require.Contains(t, out, "examples:")
	require.NotContains(t, out, "example: 7")
}

func Test_UpgradeCmdAlreadyCurrent(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff upgrade ../data/upgrade/already-31.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "openapi: 3.2.1")
	require.Contains(t, out, "title: already-31")
}

func Test_UpgradeCmdJsonFormat(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff upgrade ../data/upgrade/simple-30.yaml --format json"), &stdout, io.Discard))
	var doc map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &doc))
	require.Equal(t, "3.2.1", doc["openapi"])
}

func Test_UpgradeCmdInvalid(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 102, internal.Run(cmdToArgs("oasdiff upgrade ../data/upgrade/invalid.yaml"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), "failed to load original spec")
}

func Test_AutoUpgradeBreaking_CrossVersionEquivalent(t *testing.T) {
	// simple-30.yaml and equivalent-31.yaml describe the same API; the latter
	// is the 3.1 canonical form of the former (nullable -> type array, etc).
	// With --auto-upgrade, both specs are canonicalised so the dialect diff
	// disappears and oasdiff reports no changes.
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/upgrade/simple-30.yaml ../data/upgrade/equivalent-31.yaml --auto-upgrade --fail-on ERR"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "No changes detected")
}

func Test_AutoUpgradeChangelog_CrossVersionEquivalent(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/upgrade/simple-30.yaml ../data/upgrade/equivalent-31.yaml --auto-upgrade --fail-on ERR"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "No changes detected")
}

func Test_AutoUpgradeDiff_SameVersionIsHarmless(t *testing.T) {
	// When both specs are already on the same version, --auto-upgrade still
	// canonicalises them (idempotent). The diff result must not change.
	withoutFlag := bytes.Buffer{}
	require.Zero(t, internal.Run(cmdToArgs("oasdiff summary ../data/upgrade/already-31.yaml ../data/upgrade/already-31.yaml"), &withoutFlag, io.Discard))
	withFlag := bytes.Buffer{}
	require.Zero(t, internal.Run(cmdToArgs("oasdiff summary ../data/upgrade/already-31.yaml ../data/upgrade/already-31.yaml --auto-upgrade"), &withFlag, io.Discard))
	require.Equal(t, withoutFlag.String(), withFlag.String())
}
