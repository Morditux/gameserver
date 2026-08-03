package server

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// GameInfo représente un jeu listé dans l'index.
type GameInfo struct {
	Name  string
	Title string
}

// IndexHandler génère dynamiquement l'index des jeux à partir du dossier html/games.
func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	games, err := loadGames()
	if err != nil {
		http.Error(w, "impossible de charger les jeux : "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("html/index.html")
	if err != nil {
		http.Error(w, "impossible de charger le template : "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Games []GameInfo
	}{
		Title: "Gameserver — Index des jeux",
		Games: games,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "erreur de rendu : "+err.Error(), http.StatusInternalServerError)
	}
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

// loadGames parcourt html/games, extrait le titre de chaque page HTML et trie par titre.
func loadGames() ([]GameInfo, error) {
	files, err := filepath.Glob(filepath.Join("html", "games", "*.html"))
	if err != nil {
		return nil, err
	}

	games := make([]GameInfo, 0, len(files))
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".html")
		title := extractTitle(f)
		if title == "" {
			title = name
		}
		games = append(games, GameInfo{Name: name, Title: title})
	}

	sort.Slice(games, func(i, j int) bool {
		return games[i].Title < games[j].Title
	})

	return games, nil
}

var titleRegex = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// extractTitle retourne le contenu de la balise <title> du fichier, ou "" si absent.
func extractTitle(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	m := titleRegex.FindSubmatch(content)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(string(m[1]))
}
