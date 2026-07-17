package galaxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testClient(t *testing.T) *Client {
	t.Helper()
	return startAPI(t,
		map[string]string{"community.general": "11.2.0", "ansible.posix": "5.0.0"},
		map[string]string{"geerlingguy.docker": "8.0.0"},
	)
}

func statusByName(results []Result) map[string]Status {
	m := make(map[string]Status, len(results))
	for _, r := range results {
		m[r.Name] = r.Status
	}
	return m
}

func TestRun_updatesToLatest(t *testing.T) {
	dir := t.TempDir()
	path := writeReq(t, dir, `---
collections:
  - name: community.general
    version: 9.0.1
  - ansible.posix
roles:
  - name: geerlingguy.docker
`)

	results, err := Run(t.Context(), testClient(t), []string{path})
	require.NoError(t, err)

	status := statusByName(results)
	assert.Equal(t, StatusUpdated, status["community.general"])
	assert.Equal(t, StatusUpdated, status["ansible.posix"])
	assert.Equal(t, StatusUpdated, status["geerlingguy.docker"])

	out := readFile(t, path)
	assert.Contains(t, out, "version: 11.2.0")
	assert.Contains(t, out, "name: ansible.posix")
	assert.Contains(t, out, "version: 5.0.0")
	assert.Contains(t, out, "version: 8.0.0")
}

func TestRun_skipsNonGalaxyEntries(t *testing.T) {
	dir := t.TempDir()
	path := writeReq(t, dir, `---
collections:
  - name: private.collection
    type: git
    source: https://example.com/repo.git
roles:
  - name: geerlingguy.docker
    src: https://github.com/geerlingguy/ansible-role-docker
`)

	results, err := Run(t.Context(), testClient(t), []string{path})
	require.NoError(t, err)

	status := statusByName(results)
	assert.Equal(t, StatusSkipped, status["private.collection"])
	assert.Equal(t, StatusSkipped, status["geerlingguy.docker"])
	assert.Contains(t, readFile(t, path), "type: git")
}

func TestRun_preservesCommentsAndFormatting(t *testing.T) {
	dir := t.TempDir()
	path := writeReq(t, dir, `---
# managed by galaxy-update
collections:
  - name: community.general # the general collection
    version: 9.0.1
  # posix head
  - ansible.posix # posix line
`)

	_, err := Run(t.Context(), testClient(t), []string{path})
	require.NoError(t, err)

	out := readFile(t, path)
	assert.Contains(t, out, "# managed by galaxy-update")
	assert.Contains(t, out, "# the general collection")
	assert.Contains(t, out, "version: 11.2.0")
	assert.Contains(t, out, "# posix head")
	assert.Contains(t, out, "name: ansible.posix # posix line")
	assert.Equal(t, "---\n", out[:4])
}

func TestRun_unchangedFileNotRewritten(t *testing.T) {
	dir := t.TempDir()
	content := "---\ncollections:\n  - name: community.general\n    version: 11.2.0\n"
	path := writeReq(t, dir, content)

	results, err := Run(t.Context(), testClient(t), []string{path})
	require.NoError(t, err)

	assert.Equal(t, StatusCurrent, statusByName(results)["community.general"])
	assert.Equal(t, content, readFile(t, path))
}

func TestRun_errorOnUnresolvableEntry(t *testing.T) {
	dir := t.TempDir()
	path := writeReq(t, dir, "---\ncollections:\n  - name: missing.collection\n")

	results, err := Run(t.Context(), testClient(t), []string{path})
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, StatusError, results[0].Status)
	assert.Error(t, results[0].Err)
}

func TestRun_ignoresFilesWithoutEntries(t *testing.T) {
	dir := t.TempDir()
	for _, content := range []string{"", "roles: []\n", "- just.a.list\n"} {
		path := writeReq(t, dir, content)
		results, err := Run(t.Context(), testClient(t), []string{path})
		require.NoError(t, err)
		assert.Empty(t, results)
	}
}
