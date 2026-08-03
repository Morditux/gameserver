package main

import (
	"os"

	"github.com/Morditux/gameserver/server"
)

func main() {
	host := getEnv("HOST", "0.0.0.0")
	port := getEnv("PORT", "8080")

	srv := server.NewServer(host, port)

	// Patterns avec méthode : seuls GET (et HEAD, implicite) sont acceptés,
	// les autres méthodes reçoivent automatiquement un 405 Method Not Allowed.
	srv.HandleFunc("GET /", srv.IndexHandler)
	srv.HandleFunc("GET /games/", srv.GamesHandler)

	if err := srv.Start(); err != nil {
		panic(err)
	}
}

// getEnv retourne la valeur de la variable d'environnement ou une valeur par défaut.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
