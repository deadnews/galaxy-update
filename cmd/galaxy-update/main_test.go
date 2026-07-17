package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deadnews/galaxy-update/internal/galaxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	colorEnabled = false
	os.Exit(m.Run())
}

func TestResolveFiles(t *testing.T) {
	t.Run("explicit files returned as-is", func(t *testing.T) {
		files, err := resolveFiles([]string{"a.yml", "b.yml"})
		require.NoError(t, err)
		assert.Equal(t, []string{"a.yml", "b.yml"}, files)
	})

	t.Run("discovers files from current directory", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.yml"), []byte("collections: []\n"), 0o600))
		t.Chdir(dir)

		files, err := resolveFiles(nil)
		require.NoError(t, err)
		assert.Len(t, files, 1)
	})
}

func TestExitStatus(t *testing.T) {
	t.Run("passes when nothing failed", func(t *testing.T) {
		assert.NoError(t, exitStatus([]galaxy.Result{{Status: galaxy.StatusUpdated}}))
	})
	t.Run("fails on errors", func(t *testing.T) {
		assert.ErrorIs(t, exitStatus([]galaxy.Result{{Status: galaxy.StatusError}}), errHasErrors)
	})
}
