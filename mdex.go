package mdex

import (
	"github.com/jpbruinsslot/mdex/http"
	"github.com/jpbruinsslot/mdex/parser"
)

const (
	Version = "1.0.0"
	URL     = "github.com/jpbruinsslot/mdex"
)

func Generate(mdParser parser.MarkdownParser, opts ...parser.Option) error {
	p := parser.New(mdParser, opts...)
	return p.Generate()
}

func Serve(opts ...http.Option) error {
	srv, err := http.NewHTTPServer(opts...)
	if err != nil {
		return err
	}

	srv.Run()
	return nil
}

func GenerateAndServe(mdParser parser.MarkdownParser, parserOpts []parser.Option, serverOpts []http.Option) error {
	p := parser.New(mdParser, parserOpts...)
	if err := p.Generate(); err != nil {
		return err
	}

	return Serve(serverOpts...)
}
