package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleStatic(t *testing.T) {
	// Create a temporary directory for the static root
	tempDir, err := os.MkdirTemp("", "mdex-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srv, err := NewHTTPServer(WithStaticRoot(tempDir))
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/static/css/main.css", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := srv.handleStatic()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestHandleStaticRoute(t *testing.T) {
	// Create a temporary directory for the static root
	tempDir, err := os.MkdirTemp("", "mdex-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy index.html file
	indexFile := filepath.Join(tempDir, "index.html")
	if err := os.WriteFile(indexFile, []byte("Hello"), 0644); err != nil {
		t.Fatal(err)
	}

	srv, err := NewHTTPServer(WithStaticRoot(tempDir))
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := srv.handleStaticRoute()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Test with a non-existent file
	req, err = http.NewRequest("GET", "/non-existent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}
