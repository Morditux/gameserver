package server

import (
	"html/template"
	"log"
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
	games, err := s.loadGames()
	if err != nil {
		// Les détails de l'erreur (chemins internes) restent dans les logs :
		// ils ne sont pas exposés au client.
		log.Printf("index: impossible de charger les jeux : %v", err)
		http.Error(w, "impossible de charger les jeux", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(s.indexPath)
	if err != nil {
		log.Printf("index: impossible de charger le template : %v", err)
		http.Error(w, "impossible de charger le template", http.StatusInternalServerError)
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
		// La réponse est déjà partiellement écrite : on se contente de logger.
		log.Printf("index: erreur de rendu : %v", err)
	}
}

// GamesHandler sert les pages HTML du dossier html/games.
// Le nom du jeu est extrait de l'URL : /games/{nom} -> html/games/{nom}.html
func (s *Server) GamesHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/games/")
	if !validGameName(name) {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(s.gamesDir, name+".html")

	// On ne sert que des fichiers réguliers : ni répertoire (listage interdit),
	// ni lien symbolique (évite de fuiter un fichier hors de html/games).
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

// validGameName n'accepte qu'un nom de fichier simple : sans séparateur de
// chemin (ni "/" ni "\"), sans ".." (traversée de répertoire) et sans octet nul.
func validGameName(name string) bool {
	if name == "" || strings.Contains(name, "..") {
		return false
	}
	if strings.ContainsAny(name, `/\`) || strings.ContainsRune(name, 0) {
		return false
	}
	return true
}

// loadGames parcourt html/games, extrait le titre de chaque page HTML et trie par titre.
func (s *Server) loadGames() ([]GameInfo, error) {
	files, err := filepath.Glob(filepath.Join(s.gamesDir, "*.html"))
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
