package parser

type Options struct {
	RootPath   string
	OutputPath string
}

type Option func(o *Options)

func WithRootPath(path string) Option {
	return func(o *Options) {
		o.RootPath = path
	}
}

func WithOutputPath(path string) Option {
	return func(o *Options) {
		o.OutputPath = path
	}
}
