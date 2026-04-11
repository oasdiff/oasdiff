package generator_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker/generator"
	"github.com/stretchr/testify/require"
)

func WriteToFile(t *testing.T, filename string, lines []string) {
	t.Helper()

	file, err := os.Create(filename)
	require.NoError(t, err)
	defer file.Close() //nolint:errcheck
	for _, line := range lines {
		_, err = file.WriteString(line + "\n")
		require.NoError(t, err)
	}
}

func TestTreeGenerator(t *testing.T) {
	result, err := generator.Generate(generator.GetTree("tree.yaml"))
	require.NoError(t, err)
	slices.Sort(result)
	WriteToFile(t, "messages.yaml", result)
	require.Len(t, result, 263)
	badId, unique := isUninueIds(result)
	require.True(t, unique, badId)
}

func isUninueIds(messages []string) (string, bool) {
	ids := make(map[string]struct{})
	for _, message := range messages {
		id := strings.SplitAfter(message, ":")[0]
		if _, ok := ids[id]; ok {
			return id, false
		}
		ids[id] = struct{}{}
	}
	return "", true
}
