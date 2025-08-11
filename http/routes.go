package http

func (srv *HTTPServer) RegisterRoutes() {
	srv.Router.Handle("GET /static/", srv.handleStatic())
	srv.Router.Handle("GET /", Chain(srv.handleStaticRoute(), srv.Middleware...))
}
