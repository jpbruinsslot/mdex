package http

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type HTTPServer struct {
	Server     *http.Server
	Router     *http.ServeMux
	Logger     *slog.Logger
	StaticRoot string
	Middleware []Middleware
	BasicAuth  struct {
		Username string
		Password string
	}
}

func NewHTTPServer(opts ...Option) (*HTTPServer, error) {
	options := &Options{
		Port:       "8080",
		StaticRoot: "public",
	}
	for _, opt := range opts {
		opt(options)
	}

	srv := &HTTPServer{}

	srv.Logger = slog.Default().With("module", "http")
	srv.StaticRoot = options.StaticRoot

	// Validate the static root directory
	if err := srv.ValidateStaticRoot(); err != nil {
		return nil, fmt.Errorf("static root validation failed: %w", err)
	}

	// Set middleware
	srv.Middleware = []Middleware{
		srv.loggingMiddleware,
	}

	// Configure basic authentication if both username and password are provided
	if options.Username != "" && options.Password != "" {
		srv.BasicAuth.Username = options.Username
		srv.BasicAuth.Password = options.Password
		srv.Middleware = append(srv.Middleware, srv.basicAuthMiddleware)
		srv.Logger.Info("Basic authentication enabled", "username", srv.BasicAuth.Username)
	} else if options.Username != "" || options.Password != "" {
		// Return an error if only one of username or password is provided
		return nil, fmt.Errorf("basic auth requires both a username and a password")
	}

	// Create the router
	srv.Router = http.NewServeMux()
	srv.RegisterRoutes()

	// Create the server
	srv.Server = &http.Server{
		// Addr is the TCP address to listen on, ":http" if empty.
		Addr: fmt.Sprintf(":%s", options.Port),

		// ReadTimeout is the maximum duration for reading the entire
		// request, including the body.
		ReadTimeout: 5 * time.Second,

		// WriteTimeout is the maximum duration before timing out
		// writes of the response. It is reset whenever a new
		// request's header is read. Like ReadTimeout, it does not
		// let Handlers make decisions on a per-request basis.
		WriteTimeout: 10 * time.Second,

		// IdleTimeout is the maximum amount of time to wait for the
		// next request when keep-alive are enabled. If IdleTimeout
		// is zero, the value of ReadTimeout is used. If both are
		// zero, there is no timeout.
		IdleTimeout: 120 * time.Second,

		// MaxHeaderBytes controls the maximum number of bytes the
		// server will read parsing the request header's keys and
		// values, including the request line. It does not limit the
		// size of the request body. If zero, DefaultMaxHeaderBytes is used.
		MaxHeaderBytes: 1 << 20,
	}

	srv.Server.Handler = srv.Router
	return srv, nil
}

func (srv *HTTPServer) ValidateStaticRoot() error {
	info, err := os.Stat(srv.StaticRoot)
	if err != nil {
		return fmt.Errorf("static root inaccessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("static root is not a directory: %s", srv.StaticRoot)
	}
	return nil
}

func (srv *HTTPServer) Run() {
	srv.Logger.Info("Server is running", "addr", srv.Server.Addr)
	log.Fatal(srv.Server.ListenAndServe())
}
