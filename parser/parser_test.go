package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockMarkdownParser struct{}

func (m *mockMarkdownParser) Convert(markdown []byte) ([]byte, error) {
	return nil, nil
}

func (m *mockMarkdownParser) ExtractTOC(markdown []byte) ([]TOCEntry, error) {
	return nil, nil
}

func TestNew(t *testing.T) {
	// Test with no options
	p := New(&mockMarkdownParser{})
	if p.RootPath == "" {
		t.Error("Expected RootPath to be set to the current working directory, but it was empty")
	}
	if p.OutputPath != "./public" {
		t.Errorf("Expected OutputPath to be './public', but got '%s'", p.OutputPath)
	}

	// Test with options
	p = New(&mockMarkdownParser{}, WithRootPath("/tmp"), WithOutputPath("/tmp/public"))
	if p.RootPath != "/tmp" {
		t.Errorf("Expected RootPath to be '/tmp', but got '%s'", p.RootPath)
	}
	if p.OutputPath != "/tmp/public" {
		t.Errorf("Expected OutputPath to be '/tmp/public', but got '%s'", p.OutputPath)
	}
}

func TestGenerate(t *testing.T) {
	// Create a temporary directory for the root and output paths
	rootDir, err := os.MkdirTemp("", "mdex-root")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	outputDir, err := os.MkdirTemp("", "mdex-output")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	// Create a sample markdown file
	mdFile := filepath.Join(rootDir, "test.md")
	err = os.WriteFile(mdFile, []byte("# Hello World"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new parser and generate the site
	p := New(NewGoldmarkParser(), WithRootPath(rootDir), WithOutputPath(outputDir))
	err = p.Generate()
	if err != nil {
		t.Fatal(err)
	}

	// Check that the HTML file was created
	htmlFile := filepath.Join(outputDir, "test.html")
	if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
		t.Fatalf("Expected HTML file to be created at '%s', but it wasn't", htmlFile)
	}

	// Check that the index.html file was created
	indexPath := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatalf("Expected index.html file to be created at '%s', but it wasn't", indexPath)
	}

	// Check the contents of the HTML file
	html, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "<h1 id=\"hello-world\">Hello World</h1>") {
		t.Errorf("Expected HTML to contain '%s', but it didn't", "<h1>Hello World</h1>")
	}
}

func TestIsIgnored(t *testing.T) {
	p := &Parser{}
	if !p.isIgnored(".hidden") {
		t.Error("Expected '.hidden' to be ignored")
	}
	if !p.isIgnored("_ignored") {
		t.Error("Expected '_ignored' to be ignored")
	}
	if p.isIgnored("visible") {
		t.Error("Expected 'visible' not to be ignored")
	}
}

func TestIsMarkdownFile(t *testing.T) {
	p := &Parser{}
	if !p.isMarkdownFile("test.md") {
		t.Error("Expected 'test.md' to be a markdown file")
	}
	if !p.isMarkdownFile("test.markdown") {
		t.Error("Expected 'test.markdown' to be a markdown file")
	}
	if !p.isMarkdownFile("test.mkd") {
		t.Error("Expected 'test.mkd' to be a markdown file")
	}
	if p.isMarkdownFile("test.txt") {
		t.Error("Expected 'test.txt' not to be a markdown file")
	}
}

func TestGetDirectoryListing(t *testing.T) {
	// Create a temporary directory
	rootDir, err := os.MkdirTemp("", "mdex-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	// Create some files and directories
	os.Mkdir(filepath.Join(rootDir, "dir1"), 0755)
	os.Mkdir(filepath.Join(rootDir, ".hidden_dir"), 0755)
	os.Mkdir(filepath.Join(rootDir, "_ignored_dir"), 0755)
	os.WriteFile(filepath.Join(rootDir, "file1.md"), []byte(""), 0644)
	os.WriteFile(filepath.Join(rootDir, "file2.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(rootDir, ".hidden_file.md"), []byte(""), 0644)
	os.WriteFile(filepath.Join(rootDir, "_ignored_file.md"), []byte(""), 0644)

	p := New(&mockMarkdownParser{}, WithRootPath(rootDir))

	// Test listing the root directory
	entries, err := p.getDirectoryListing(rootDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := []FileEntry{
		{Name: "dir1", IsDir: true},
		{Name: "file1.md", IsDir: false},
	}

	if len(entries) != len(expected) {
		t.Fatalf("Expected %d entries, but got %d", len(expected), len(entries))
	}

	for i, entry := range entries {
		if entry.Name != expected[i].Name || entry.IsDir != expected[i].IsDir {
			t.Errorf("Expected entry %v, but got %v", expected[i], entry)
		}
	}
}
