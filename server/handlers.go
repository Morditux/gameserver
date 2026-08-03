package server

import (
	"net/http"
	"path/filepath"
	"strings"
)

// IndexHandler sert la page d'accueil index.html.
func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/index.html")
}

// GamesHandler sert les pages HTML du dossier html/games.
// Le nom du jeu est extrait de l'URL : /games/{nom} -> html/games/{nom}.html
func (s *Server) GamesHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/games/")
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("html", "games", name+".html"))
}
