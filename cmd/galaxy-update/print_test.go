package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/deadnews/galaxy-update/internal/galaxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	require.NoError(t, w.Close())
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data)
}

func TestPrintResults(t *testing.T) {
	results := []galaxy.Result{
		{File: "requirements.yml", Name: "community.general", Old: "9.0.1", New: "11.2.0", Status: galaxy.StatusUpdated},
		{File: "requirements.yml", Name: "ansible.posix", New: "5.0.0", Status: galaxy.StatusUpdated},
		{File: "requirements.yml", Name: "private.collection", Status: galaxy.StatusSkipped},
		{File: "requirements.yml", Name: "missing.collection", Status: galaxy.StatusError, Err: errors.New("404 not found")},
	}

	t.Run("shows changes on one line, hides skipped", func(t *testing.T) {
		out := captureStdout(t, func() { printResults(results, false) })
		assert.Contains(t, out, "requirements.yml")
		assert.Contains(t, out, "UPDATED")
		assert.Contains(t, out, "community.general 9.0.1 → 11.2.0")
		assert.Contains(t, out, "ansible.posix → 5.0.0")
		assert.Contains(t, out, "404 not found")
		assert.NotContains(t, out, "SKIP")
	})

	t.Run("verbose shows skipped", func(t *testing.T) {
		out := captureStdout(t, func() { printResults(results, true) })
		assert.Contains(t, out, "SKIP")
	})
}

func TestPrintResults_allCurrentHiddenWithoutVerbose(t *testing.T) {
	results := []galaxy.Result{{File: "requirements.yml", Name: "community.general", Old: "11.2.0", Status: galaxy.StatusCurrent}}
	out := captureStdout(t, func() { printResults(results, false) })
	assert.Empty(t, out)
}

func TestPrintResults_groupsFilesWithBlankSeparator(t *testing.T) {
	results := []galaxy.Result{
		{File: "a.yml", Name: "ns.one", New: "1.0.0", Status: galaxy.StatusUpdated},
		{File: "a.yml", Name: "ns.skip", Status: galaxy.StatusSkipped},
		{File: "b.yml", Name: "ns.two", New: "2.0.0", Status: galaxy.StatusUpdated},
	}
	out := captureStdout(t, func() { printResults(results, false) })

	lines := strings.Split(out, "\n")
	require.Equal(t, "a.yml", lines[0])
	assert.Contains(t, lines[1], "ns.one")
	assert.Empty(t, lines[2], "blank line separates file groups")
	assert.Equal(t, "b.yml", lines[3])
	assert.Contains(t, lines[4], "ns.two")
	assert.Empty(t, lines[5], "trailing blank line after last group")
	assert.NotContains(t, out, "ns.skip", "hidden entries do not appear")
}
