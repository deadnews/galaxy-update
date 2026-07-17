package galaxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const collectionIndexPrefix = "/v3/plugin/ansible/content/published/collections/index/"

func startAPI(t *testing.T, collections, roles map[string]string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(collectionIndexPrefix, func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, collectionIndexPrefix), "/"), "/")
		version, ok := collections[parts[0]+"."+parts[1]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		writeJSON(t, w, map[string]any{"highest_version": map[string]string{"version": version}})
	})
	mux.HandleFunc("/v1/roles/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("namespace") + "." + r.URL.Query().Get("name")
		version, ok := roles[name]
		if !ok {
			writeJSON(t, w, map[string]any{"results": []any{}})
			return
		}
		writeJSON(t, w, map[string]any{"results": []any{roleResult(version)}})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Client{http: srv.Client(), baseURL: srv.URL}
}

func roleResult(versions ...string) map[string]any {
	list := make([]any, len(versions))
	for i, v := range versions {
		list[i] = map[string]string{"name": v}
	}
	return map[string]any{"summary_fields": map[string]any{"versions": list}}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	_ = json.NewEncoder(w).Encode(v)
}

func writeReq(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "requirements.yml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}
