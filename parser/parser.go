package parser

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jpbruinsslot/mdex/templates"
)

type MarkdownParser interface {
	Convert(markdown []byte) ([]byte, error)
	ExtractTOC(markdown []byte) ([]TOCEntry, error)
}

type Parser struct {
	Logger     *slog.Logger
	Templates  map[string]*template.Template
	RootPath   string
	OutputPath string
	Parser     MarkdownParser
}

type TemplateData struct {
	Content template.HTML
	Title   string
	Files   []FileEntry
	TOC     []TOCEntry
	IsIndex bool
}

type FileEntry struct {
	Name      string
	IsDir     bool
	Path      string
	RoutePath string
}

type TOCEntry struct {
	Title string
	Level int
	ID    string
}

func New(mdParser MarkdownParser, opts ...Option) *Parser {
	options := &Options{
		RootPath:   "",
		OutputPath: "./public",
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.RootPath == "" {
		currentWd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get current working directory: %v", err)
		}
		options.RootPath = currentWd
	}

	p := &Parser{
		Logger:     slog.Default(),
		Templates:  make(map[string]*template.Template),
		RootPath:   options.RootPath,
		OutputPath: options.OutputPath,
		Parser:     mdParser,
	}
	p.loadEmbeddedTemplates()

	return p
}

func (p *Parser) loadEmbeddedTemplates() {
	templateFiles, err := templates.FS.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range templateFiles {
		name := file.Name()

		if file.IsDir() || !strings.HasSuffix(name, ".html") || name == "base.html" {
			continue
		}

		// Parse base.html + the current file together
		tmpl, err := template.ParseFS(templates.FS, "base.html", "sidebar.html", "toc.html", name)
		if err != nil {
			log.Fatalf("failed to parse template %s: %v", name, err)
		}

		key := strings.TrimSuffix(name, ".html")
		p.Templates[key] = tmpl
	}
}

func (p *Parser) renderTemplate(name string, data TemplateData) (string, error) {
	tmpl, ok := p.Templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (p *Parser) getDirectoryListing(root string) ([]FileEntry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var result []FileEntry

	// Add .. to go up one directory level
	if root != p.RootPath {
		result = append(result, FileEntry{
			Name:      "..",
			IsDir:     true,
			Path:      filepath.Join(root, ".."),
			RoutePath: strings.TrimSuffix(filepath.Join(root, ".."), ".md"),
		})
	}

	for _, entry := range entries {
		// Skip the output directory
		if entry.Name() == filepath.Base(p.OutputPath) {
			continue
		}

		// Skip hidden files and folder, but allow directories, and markdown files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip files and folders that start with an underscore
		if strings.HasPrefix(entry.Name(), "_") {
			continue
		}

		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".md") {
			result = append(result, FileEntry{
				Name:      entry.Name(),
				IsDir:     entry.IsDir(),
				Path:      filepath.Join(root, entry.Name()),
				RoutePath: strings.TrimSuffix(filepath.Join(root, entry.Name()), ".md"),
			})
		}
	}

	// Sort: folders first, then files, both alphabetically
	slices.SortFunc(result, func(a, b FileEntry) int {
		if a.IsDir && !b.IsDir {
			return -1
		}
		if !a.IsDir && b.IsDir {
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})

	return result, nil

}

// Ignore files and directories that start with an underscore or a dot
func (p *Parser) isIgnored(name string) bool {
	return strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".")
}

// Check if the file has a .md extension
func (p *Parser) isMarkdownFile(name string) bool {
	return strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".markdown") || strings.HasSuffix(name, ".mkd")
}

func (p *Parser) ensureIndexForDir(dir string, files []FileEntry) error {
	indexPath := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		return nil // already exists
	}

	content := template.HTML("")
	toc := []TOCEntry{}

	// If there is an index.md file, we use that to generate the index
	indexMDPath := filepath.Join(dir, "index.md")
	if _, err := os.Stat(indexMDPath); err == nil {
		markdown, err := os.ReadFile(indexMDPath)
		if err != nil {
			return fmt.Errorf("failed to read index.md: %w", err)
		}

		html, err := p.Parser.Convert(markdown)
		if err != nil {
			return fmt.Errorf("failed to convert index.md to HTML: %w", err)
		}

		toc, err = p.Parser.ExtractTOC(markdown)
		if err != nil {
			return fmt.Errorf("failed to extract TOC from index.md: %w", err)
		}

		content = template.HTML(html)
	}

	indexData := TemplateData{
		Title:   "Index of " + filepath.Base(dir),
		Content: content,
		Files:   files,
		TOC:     toc,
		IsIndex: true,
	}

	rendered, err := p.renderTemplate("index", indexData)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(p.OutputPath, strings.TrimPrefix(dir, p.RootPath), "index.html")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return p.Save(rendered, outputPath)
}

func (p *Parser) Generate() error {
	return filepath.WalkDir(p.RootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip the output directory
			if path == p.OutputPath {
				return nil
			}

			// Unless it is the root directory, we skip directories that are ignored
			if path != p.RootPath && p.isIgnored(d.Name()) {
				return filepath.SkipDir
			}

			files, err := p.getDirectoryListing(path)
			if err != nil {
				return err
			}

			if err := p.ensureIndexForDir(path, files); err != nil {
				return err
			}
			return nil
		}

		if !p.isMarkdownFile(d.Name()) {
			return nil
		}

		if p.isIgnored(d.Name()) {
			return nil
		}

		p.Logger.Info("Processing", "file", path)

		markdown, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		html, err := p.Parser.Convert(markdown)
		if err != nil {
			return err
		}

		toc, err := p.Parser.ExtractTOC(markdown)
		if err != nil {
			return err
		}

		parentDir := filepath.Dir(path)
		files, err := p.getDirectoryListing(parentDir)
		if err != nil {
			return err
		}

		title := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		data := TemplateData{
			Content: template.HTML(html),
			Title:   title,
			Files:   files,
			TOC:     toc,
			IsIndex: false,
		}

		relPath, err := filepath.Rel(p.RootPath, path)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(p.OutputPath, strings.TrimSuffix(relPath, filepath.Ext(relPath))+".html")
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		rendered, err := p.renderTemplate("single", data)
		if err != nil {
			return err
		}

		return p.Save(rendered, outputPath)
	})
}

func (p *Parser) Save(html, outputFilePath string) error {
	// Save the HTML to the output path
	if err := os.MkdirAll(filepath.Dir(outputFilePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(outputFilePath, []byte(html), 0644)
}
