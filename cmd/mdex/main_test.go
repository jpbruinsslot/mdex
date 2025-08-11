package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mdex "github.com/jpbruinsslot/mdex"
	"github.com/jpbruinsslot/mdex/parser"
)

func TestRunGenerate(t *testing.T) {
	// Create temporary directories for root and output
	rootDir, err := os.MkdirTemp("", "mdex-cmd-root")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	outputDir, err := os.MkdirTemp("", "mdex-cmd-output")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	// Create a sample markdown file
	mdFile := filepath.Join(rootDir, "test.md")
	err = os.WriteFile(mdFile, []byte("# Command Test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Run the generate command
	err = run([]string{"generate", "--root", rootDir, "--output", outputDir})
	if err != nil {
		t.Fatalf("run generate command failed: %v", err)
	}

	// Verify the HTML file was created
	htmlFile := filepath.Join(outputDir, "test.html")
	if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
		t.Fatalf("Expected HTML file to be created at '%s', but it wasn't", htmlFile)
	}

	// Verify content
	htmlContent, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(htmlContent), "<h1 id=\"command-test\">Command Test</h1>") {
		t.Errorf("Expected HTML content to contain heading, but it didn't: %s", string(htmlContent))
	}
}

func TestRunServe(t *testing.T) {
	// Create temporary directories for root and output
	rootDir, err := os.MkdirTemp("", "mdex-cmd-serve-root")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	outputDir, err := os.MkdirTemp("", "mdex-cmd-serve-output")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	// Create a sample markdown file
	mdFile := filepath.Join(rootDir, "test-serve.md")
	err = os.WriteFile(mdFile, []byte("# Serve Test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Generate the site first
	mdParser := parser.NewGoldmarkParser()
	parserOpts := []parser.Option{
		parser.WithRootPath(rootDir),
		parser.WithOutputPath(outputDir),
	}
	err = mdex.Generate(mdParser, parserOpts...)
	if err != nil {
		t.Fatalf("mdex.Generate failed: %v", err)
	}

	// Capture stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Run the serve command in a goroutine
	done := make(chan error)
	go func() {
		done <- run([]string{"serve", "--static-root", outputDir, "--port", "8081"})
	}()

	// Restore stdout/stderr
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Give the server some time to start
	time.Sleep(2 * time.Second)

	// Make an HTTP request to the server
	resp, err := http.Get("http://127.0.0.1:8081/test-serve.html")
	if err != nil {
		t.Fatalf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Stderr: %s", http.StatusOK, resp.StatusCode, string(out))
	}

	// Verify content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "<h1 id=\"serve-test\">Serve Test</h1>") {
		t.Errorf("Expected body to contain heading, but it didn't: %s", string(body))
	}

	// Check if the serve command returned an error (it should block until killed)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve command exited with error: %v", err)
		}
	default:
		// Server is still running, which is expected
	}
}
