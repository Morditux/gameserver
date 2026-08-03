package server

import (
	"net/http"
	"time"
)

// Simple http server implementation
type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	port       string
	listenAddr string
	gamesDir   string // répertoire contenant les pages de jeux
	indexPath  string // chemin du template d'index
}

func NewServer(addr string, port string) *Server {
	listenAddr := addr + ":" + port
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:              listenAddr,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second, // protège contre les connexions lentes (slowloris)
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
	}
	return &Server{
		httpServer: httpServer,
		mux:        mux,
		port:       port,
		listenAddr: listenAddr,
		gamesDir:   "html/games",
		indexPath:  "html/index.html",
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

// securityHeaders ajoute des en-têtes de sécurité à toutes les réponses HTTP.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}
