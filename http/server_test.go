package http

import (
	"os"
	"strings"
	"testing"
)

func TestNewHTTPServer(t *testing.T) {
	// Create a temporary directory for the static root
	tempDir, err := os.MkdirTemp("", "mdex-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy public directory for the default case
	publicDir := "public"
	if err := os.Mkdir(publicDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(publicDir)

	// Test with no options
	srv, err := NewHTTPServer(WithStaticRoot(publicDir))
	if err != nil {
		t.Fatal(err)
	}
	if srv.Server.Addr != ":8080" {
		t.Errorf("Expected Addr to be ':8080', but got '%s'", srv.Server.Addr)
	}
	if srv.StaticRoot != "public" {
		t.Errorf("Expected StaticRoot to be 'public', but got '%s'", srv.StaticRoot)
	}
	if srv.BasicAuth.Username != "" || srv.BasicAuth.Password != "" {
		t.Errorf("Expected BasicAuth to be empty, but got username '%s' and password '%s'", srv.BasicAuth.Username, srv.BasicAuth.Password)
	}

	// Test with options
	srv, err = NewHTTPServer(WithPort("8888"), WithStaticRoot(tempDir), WithBasicAuth("user", "pass"))
	if err != nil {
		t.Fatal(err)
	}
	if srv.Server.Addr != ":8888" {
		t.Errorf("Expected Addr to be ':8888', but got '%s'", srv.Server.Addr)
	}
	if srv.StaticRoot != tempDir {
		t.Errorf("Expected StaticRoot to be '%s', but got '%s'", tempDir, srv.StaticRoot)
	}
	if srv.BasicAuth.Username != "user" {
		t.Errorf("Expected Username to be 'user', but got '%s'", srv.BasicAuth.Username)
	}
	if srv.BasicAuth.Password != "pass" {
		t.Errorf("Expected Password to be 'pass', but got '%s'", srv.BasicAuth.Password)
	}
}

func TestNewHTTPServer_IncompleteBasicAuth(t *testing.T) {
	// Create a temporary directory for the static root
	tempDir, err := os.MkdirTemp("", "mdex-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test with username but no password
	_, err = NewHTTPServer(WithStaticRoot(tempDir), WithBasicAuth("user", ""))
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	expected := "basic auth requires both a username and a password"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error to contain \"%s\", but got \"%s\"", expected, err.Error())
	}

	// Test with password but no username
	_, err = NewHTTPServer(WithStaticRoot(tempDir), WithBasicAuth("", "pass"))
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error to contain \"%s\", but got \"%s\"", expected, err.Error())
	}
}