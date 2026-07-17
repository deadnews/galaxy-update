package galaxy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRequirementsFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"requirements.yml", true},
		{"requirements.yaml", true},
		{"path/to/requirements.yml", true},
		{"Requirements.YML", true},
		{"requirements.txt", false},
		{"galaxy.yml", false},
		{"main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, isRequirementsFile(tt.path))
		})
	}
}

func TestDiscoverFiles(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "roles")
	require.NoError(t, os.MkdirAll(nested, 0o750))
	dotDir := filepath.Join(root, ".git")
	require.NoError(t, os.MkdirAll(dotDir, 0o750))

	write := func(dir, name string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("collections: []\n"), 0o600))
	}
	write(root, "requirements.yml")
	write(nested, "requirements.yaml")
	write(root, "README.md")
	write(dotDir, "requirements.yml")

	files, err := DiscoverFiles(root)
	require.NoError(t, err)
	assert.Len(t, files, 2, "finds requirements files, skips dot-dirs and other files")
}

func TestDiscoverFiles_invalidRoot(t *testing.T) {
	_, err := DiscoverFiles(filepath.Join(t.TempDir(), "nonexistent"))
	assert.Error(t, err)
}
