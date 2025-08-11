package http

type Options struct {
	Port       string
	StaticRoot string
	Username   string
	Password   string
}

type Option func(*Options)

func WithPort(port string) Option {
	return func(o *Options) {
		o.Port = port
	}
}

func WithStaticRoot(path string) Option {
	return func(o *Options) {
		o.StaticRoot = path
	}
}

func WithBasicAuth(user, pass string) Option {
	return func(o *Options) {
		o.Username = user
		o.Password = pass
	}
}
