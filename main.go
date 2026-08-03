package main

import (
	"os"

	"github.com/Morditux/gameserver/server"
)

func main() {
	host := getEnv("HOST", "0.0.0.0")
	port := getEnv("PORT", "8080")

	srv := server.NewServer(host, port)

	srv.HandleFunc("/", srv.IndexHandler)
	srv.HandleFunc("/games/", srv.GamesHandler)

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
