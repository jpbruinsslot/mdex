package mdex_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jpbruinsslot/mdex"
	mdexhttp "github.com/jpbruinsslot/mdex/http"
	"github.com/jpbruinsslot/mdex/parser"
)

func TestGenerate(t *testing.T) {
	// Create a temporary directory for the root and output paths
	rootDir, err := os.MkdirTemp("", "mdex-root-integration")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	outputDir, err := os.MkdirTemp("", "mdex-output-integration")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	// Create a sample markdown file
	mdFile := filepath.Join(rootDir, "test.md")
	err = os.WriteFile(mdFile, []byte("# Hello Integration Test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Define parser and options
	mdParser := parser.NewGoldmarkParser()
	parserOpts := []parser.Option{
		parser.WithRootPath(rootDir),
		parser.WithOutputPath(outputDir),
	}

	// Run the Generate function
	err = mdex.Generate(mdParser, parserOpts...)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Check that the HTML file was created
	htmlFile := filepath.Join(outputDir, "test.html")
	if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
		t.Fatalf("Expected HTML file to be created at '%s', but it wasn't", htmlFile)
	}

	// Check the contents of the HTML file
	html, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Fatal(err)
	}

	expectedHTML := "<h1 id=\"hello-integration-test\">Hello Integration Test</h1>"
	if !strings.Contains(string(html), expectedHTML) {
		t.Errorf("Expected HTML to contain '%s', but it didn't", expectedHTML)
	}
}

func TestGenerateAndServe(t *testing.T) {
	// Create a temporary directory for the root and output paths
	rootDir, err := os.MkdirTemp("", "mdex-root-e2e")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	outputDir, err := os.MkdirTemp("", "mdex-output-e2e")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	// Create a sample markdown file
	mdFile := filepath.Join(rootDir, "test.md")
	err = os.WriteFile(mdFile, []byte("# Hello E2E Test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Generate the site
	mdParser := parser.NewGoldmarkParser()
	parserOpts := []parser.Option{
		parser.WithRootPath(rootDir),
		parser.WithOutputPath(outputDir),
	}
	err = mdex.Generate(mdParser, parserOpts...)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Create a new HTTP server
	httpServer, err := mdexhttp.NewHTTPServer(mdexhttp.WithStaticRoot(outputDir))
	if err != nil {
		t.Fatal(err)
	}
	testServer := httptest.NewServer(httpServer.Router)
	defer testServer.Close()

	// Make a request to the server
	resp, err := http.Get(testServer.URL + "/test.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	// Check the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectedBody := "<h1 id=\"hello-e2e-test\">Hello E2E Test</h1>"
	if !strings.Contains(string(body), expectedBody) {
		t.Errorf("Expected body to contain '%s', but it didn't", expectedBody)
	}
}
