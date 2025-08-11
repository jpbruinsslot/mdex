package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jpbruinsslot/mdex/http/assets"
)

func (srv *HTTPServer) handleStatic() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staticHandler := http.FileServer(http.FS(assets.FS))
		http.StripPrefix("/static/", staticHandler).ServeHTTP(w, r)
	})
}

func (srv *HTTPServer) handleStaticRoute() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// When it has a trailing slash try the index.html page
		if strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/") + "/index"
		}

		// Resolve the full path to the requested file
		requestedPath := filepath.Join(srv.StaticRoot, filepath.Clean(path))

		// If the path does not end with .html, add it
		if !strings.HasSuffix(requestedPath, ".html") {
			requestedPath += ".html"
		}

		// Get absolute paths
		absStaticRoot, err := filepath.Abs(srv.StaticRoot)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		absRequestedPath, err := filepath.Abs(requestedPath)
		if err != nil {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}

		// Prevent path traversal: ensure requested file is within static root
		if !strings.HasPrefix(absRequestedPath, absStaticRoot) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check if file exists
		if _, err := os.Stat(absRequestedPath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		// Serve the file
		http.ServeFile(w, r, absRequestedPath)
	})
}
