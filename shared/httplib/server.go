package httplib

import (
	"fmt"
	"net/http"
)

type (
	MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc
	HandlerFunc    func(*http.Request) (any, error)
)

type Server struct {
	pathPrefix  string
	middlewares []MiddlewareFunc
	server      *http.ServeMux
}

func NewServer() *Server {
	return &Server{
		pathPrefix:  "",
		middlewares: nil,
		server:      http.NewServeMux(),
	}
}

func (s *Server) GET(path string, handler HandlerFunc) *Server {
	return s.Handle("GET", path, handler)
}

func (s *Server) POST(path string, handler HandlerFunc) *Server {
	return s.Handle("POST", path, handler)
}

func (s *Server) PUT(path string, handler HandlerFunc) *Server {
	return s.Handle("PUT", path, handler)
}

func (s *Server) DELETE(path string, handler HandlerFunc) *Server {
	return s.Handle("DELETE", path, handler)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

func (s *Server) Mount(prefix string) *Server {
	cpMiddlewares := make([]MiddlewareFunc, len(s.middlewares))
	copy(cpMiddlewares, s.middlewares)

	return &Server{
		pathPrefix:  s.pathPrefix + prefix,
		middlewares: cpMiddlewares,
		server:      s.server,
	}
}

func (s *Server) Handle(method, path string, handler HandlerFunc) *Server {
	f := HandlerJSON(handler)
	for _, m := range s.middlewares {
		f = m(f)
	}

	pattern := fmt.Sprintf("%s %s%s", method, s.pathPrefix, path)
	s.server.Handle(pattern, f)

	return s
}

func (s *Server) Use(m ...MiddlewareFunc) *Server {
	s.middlewares = append(s.middlewares, m...)
	return s
}
