package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	mdex "github.com/jpbruinsslot/mdex"
	"github.com/jpbruinsslot/mdex/http"
	"github.com/jpbruinsslot/mdex/parser"
)

const usageText = `NAME:
    mdex - markdown static site generator

USAGE:
    mdex [command] [options]

VERSION:
    %s

WEBSITE:
    %s

COMMANDS:
    generate       generate static files
    serve          serve static files
    help           show this help message

OPTIONS FOR "generate":
    --parser       parser to use (default: goldmark)
    --root         root path for markdown files (default: current directory)
    --output       output path for generated files (default: ./public)

OPTIONS FOR "serve":
    --static-root  root path to serve (default: ./public)
    --port         port to serve on (default: 8080)
    --basic-auth   username:password for basic auth (optional)
`

var errShowUsage = fmt.Errorf("show usage")

type commonFlags struct {
	parserName *string
	root       *string
	output     *string
	staticRoot *string
	port       *string
	basicAuth  *string
}

func (cf *commonFlags) parse(fs *flag.FlagSet, args []string) error {
	cf.parserName = fs.String("parser", "goldmark", "parser to use")
	cf.root = fs.String("root", ".", "root path for markdown files")
	cf.output = fs.String("output", "./public", "output path for generated files")
	cf.staticRoot = fs.String("static-root", "./public", "path to serve")
	cf.port = fs.String("port", "8080", "port to serve on")
	cf.basicAuth = fs.String("basic-auth", "", "username:password for basic auth")
	return fs.Parse(args)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if err == errShowUsage {
			fmt.Printf(usageText, mdex.Version, mdex.URL)
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return defaultCmd(args)
	}

	subcommand := args[0]
	subcommandArgs := args[1:]

	switch subcommand {
	case "generate":
		return generateCmd(subcommandArgs)
	case "serve":
		return serveCmd(subcommandArgs)
	case "help":
		return errShowUsage
	default:
		// Default behavior if no subcommand is provided
		return defaultCmd(args)
	}
}

func generateCmd(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	cf := &commonFlags{}
	if err := cf.parse(fs, args); err != nil {
		return err
	}

	md := getMarkdownParser(*cf.parserName)

	err := mdex.Generate(md, parser.WithRootPath(*cf.root), parser.WithOutputPath(*cf.output))
	if err != nil {
		return fmt.Errorf("failed to generate: %w", err)
	}
	return nil
}

func serveCmd(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	cf := &commonFlags{}
	if err := cf.parse(fs, args); err != nil {
		return err
	}

	options := []http.Option{
		http.WithStaticRoot(*cf.staticRoot),
		http.WithPort(*cf.port),
	}

	if *cf.basicAuth != "" {
		username, password, err := parseBasicAuth(*cf.basicAuth)
		if err != nil {
			return fmt.Errorf("invalid basic auth format: %w", err)
		}
		options = append(options, http.WithBasicAuth(username, password))
	}

	return mdex.Serve(options...)
}

func defaultCmd(args []string) error {
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	cf := &commonFlags{}
	if err := cf.parse(fs, args); err != nil {
		return err
	}

	md := getMarkdownParser(*cf.parserName)

	parserOpts := []parser.Option{
		parser.WithRootPath(*cf.root),
		parser.WithOutputPath(*cf.output),
	}

	serverOpts := []http.Option{
		http.WithStaticRoot(*cf.output),
		http.WithPort(*cf.port),
	}

	if *cf.basicAuth != "" {
		username, password, err := parseBasicAuth(*cf.basicAuth)
		if err != nil {
			return fmt.Errorf("invalid basic auth format: %w", err)
		}
		serverOpts = append(serverOpts, http.WithBasicAuth(username, password))
	}

	return mdex.GenerateAndServe(md, parserOpts, serverOpts)
}

func parseBasicAuth(basicAuth string) (string, string, error) {
	parts := strings.SplitN(basicAuth, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("basic-auth must be in the format username:password")
	}
	return parts[0], parts[1], nil
}

func getMarkdownParser(parserName string) parser.MarkdownParser {
	switch parserName {
	case "goldmark":
		return parser.NewGoldmarkParser()
	default:
		return parser.NewGoldmarkParser()
	}
}
