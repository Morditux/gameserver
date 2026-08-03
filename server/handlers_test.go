package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestServer construit un serveur dont les données pointent vers testdata.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	s := NewServer("127.0.0.1", "0")
	s.gamesDir = filepath.Join("testdata", "games")
	s.indexPath = filepath.Join("testdata", "index.html")
	return s
}

func TestGamesHandlerServe(t *testing.T) {
	s := newTestServer(t)
	rec := httptest.NewRecorder()
	s.GamesHandler(rec, httptest.NewRequest(http.MethodGet, "/games/game1", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Game 1") {
		t.Fatalf("corps inattendu : %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", ct)
	}
}

func TestGamesHandlerTraversal(t *testing.T) {
	s := newTestServer(t)
	cases := []string{
		"/games/../../etc/passwd",         // traversée simple
		"/games/..%2f..%2fetc%2fpasswd",   // séparateurs encodés
		"/games/%2e%2e/%2e%2e/etc/passwd", // points encodés
		"/games/..",                       // point-point seul
		"/games/a/../b",                   // segment interne
		"/games/game1.html",               // le handler ajoute .html : 404 attendu
		"/games/%00",                      // octet nul encodé
		"/games/..\\..\\etc\\passwd",      // séparateurs Windows (défense en profondeur)
	}
	for _, u := range cases {
		rec := httptest.NewRecorder()
		s.GamesHandler(rec, httptest.NewRequest(http.MethodGet, u, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s : code = %d, want 404", u, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "root:") {
			t.Errorf("%s : contenu sensible fuité : %q", u, rec.Body.String())
		}
	}
}

func TestGamesHandlerRejectsDirectory(t *testing.T) {
	s := newTestServer(t)
	dir := filepath.Join(s.gamesDir, "dir.html")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(dir) })

	rec := httptest.NewRecorder()
	s.GamesHandler(rec, httptest.NewRequest(http.MethodGet, "/games/dir", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404 (pas de listage de répertoire)", rec.Code)
	}
}

func TestGamesHandlerRejectsSymlink(t *testing.T) {
	s := newTestServer(t)
	target := filepath.Join("testdata", "outside.txt")
	if err := os.WriteFile(target, []byte("TOP SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(target) })

	link := filepath.Join(s.gamesDir, "link.html")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks non supportés : %v", err)
	}
	t.Cleanup(func() { os.Remove(link) })

	rec := httptest.NewRecorder()
	s.GamesHandler(rec, httptest.NewRequest(http.MethodGet, "/games/link", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404 (lien symbolique refusé)", rec.Code)
	}
}

func TestIndexHandler(t *testing.T) {
	s := newTestServer(t)
	rec := httptest.NewRecorder()
	s.IndexHandler(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "game1") {
		t.Fatalf("le jeu n'apparaît pas dans l'index : %q", rec.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, k := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Content-Security-Policy",
	} {
		if rec.Header().Get(k) == "" {
			t.Errorf("en-tête de sécurité %s manquant", k)
		}
	}
}

// TestMuxMethodPatterns vérifie que le mux n'accepte que GET/HEAD.
func TestMuxMethodPatterns(t *testing.T) {
	s := newTestServer(t)
	s.mux.HandleFunc("GET /", s.IndexHandler)
	s.mux.HandleFunc("GET /games/", s.GamesHandler)

	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/games/game1", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /games/game1 : code = %d, want 405", rec.Code)
	}

	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/games/game1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /games/game1 : code = %d, want 200", rec.Code)
	}
}
