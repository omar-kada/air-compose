// Package server provides implementations of http and ws handlers.
package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSPAHandler_ServesFile(t *testing.T) {
	// Create a temp directory with an index.html
	dir := t.TempDir()
	content := "<html><body>Hello</body></html>"
	err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(content), 0644)
	assert.NoError(t, err)

	// Create a subdirectory with a file
	subDir := filepath.Join(dir, "app")
	err = os.Mkdir(subDir, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "index.html"), []byte("<html><body>App</body></html>"), 0644)
	assert.NoError(t, err)

	// Create a static asset file
	err = os.WriteFile(filepath.Join(dir, "style.css"), []byte("body { color: red; }"), 0644)
	assert.NoError(t, err)

	handler := newSPAHandler(dir)

	t.Run("serves existing file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/style.css", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "body { color: red; }")
	})

	t.Run("serves index.html for nonexistent path (SPA fallback)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/dashboard", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello")
	})

	t.Run("serves index.html for root path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello")
	})

	t.Run("serves directory index.html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/app/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "App")
	})

	t.Run("serves directory index.html without trailing slash", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/app", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestNewSPAHandler_PathTraversalBlocked(t *testing.T) {
	dir := t.TempDir()
	content := "<html><body>Hello</body></html>"
	err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(content), 0644)
	assert.NoError(t, err)

	// Create a secret file outside the front directory
	secretPath := filepath.Join(os.TempDir(), "spa_secret.txt")
	err = os.WriteFile(secretPath, []byte("secret data"), 0644)
	assert.NoError(t, err)
	defer os.Remove(secretPath)

	handler := newSPAHandler(dir)

	t.Run("blocks path traversal with ../", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/../../../etc/passwd", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid path")
	})

	t.Run("blocks path traversal to parent directory", func(t *testing.T) {
		// Compute relative path from dir to the secret file in os.TempDir()
		relPath, _ := filepath.Rel(dir, secretPath)
		traversalPath := "/" + filepath.ToSlash(relPath)

		req := httptest.NewRequest("GET", traversalPath, nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		// The secret content should NOT be served
		assert.NotContains(t, w.Body.String(), "secret data")
	})
}

func TestNewSPAHandler_NoIndexFile(t *testing.T) {
	// Directory with no index.html should return 404 or error
	dir := t.TempDir()
	handler := newSPAHandler(dir)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	// Should try to serve index.html, but it doesn't exist
	// filepath.Join will return the path, os.Stat will find "not found"
	// So it should call http.ServeFile with a non-existent path
	// ServeFile will return 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}
