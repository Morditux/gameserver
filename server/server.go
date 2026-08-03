package server

import "net/http"

// Simple http server implementation
type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	port       string
	listenAddr string
}

func NewServer(addr string, port string) *Server {
	listenAddr := addr + ":" + port
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
	return &Server{
		httpServer: httpServer,
		mux:        mux,
		port:       port,
		listenAddr: listenAddr,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.httpServer.Close()
}

func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}
