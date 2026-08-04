# AGENTS.md

## Vue d'ensemble
Serveur HTTP Go minimal qui sert un index de jeux HTML5. Chaque jeu est un **fichier HTML autonome** (JS vanilla + Canvas, aucune dépendance externe). UI, commentaires et messages de commit en **français**.

## Commandes
- `go build .` — compilation
- `go run .` — serveur sur `:8080` (variables d'env `HOST`, `PORT`)
- `go test ./...` — tests (ils utilisent `server/testdata/`, pas `html/`)
- `go vet ./...` — analyse statique
- Docker : `docker build -t gameserver .` — build multi-étapes, **pas d'EXPOSE** (serveur derrière un reverse proxy), exécution en utilisateur non-root

## Architecture
- `main.go` — point d'entrée, enregistre les routes `GET /` et `GET /games/` (toute autre méthode → 405 automatique via le `ServeMux` 1.22+)
- `server/server.go` — `Server` (`http.Server` durci : timeouts anti-slowloris, `MaxHeaderBytes`), middleware `securityHeaders` (CSP, X-Frame-Options, nosniff…)
- `server/handlers.go` — `IndexHandler` (index dynamique), `GamesHandler` (sert `html/games/{nom}.html`, protégé par `validGameName` contre la traversée), `loadGames`/`extractTitle`
- `html/index.html` — template Go (`{{ .Title }}`, `{{ range .Games }}`)
- `html/games/*.html` — un jeu par fichier

## Ajouter un jeu
1. Créer `html/games/{nom}.html` avec une balise `<title>` (c'est le titre affiché dans l'index).
2. Il est servi automatiquement sur `/games/{nom}` — **aucune modification Go requise**.
3. S'inspirer des jeux existants (`snake.html`, `galaga.html`, `chasseur.html`).

## Conventions
- Commentaires, UI, messages de commit en **français** (préfixes de commit : `feat:`, `fix:`).
- Jeux : aucun JS externe, pas de build, sons WebAudio, records en `localStorage`, JS dans une IIFE, thème néon sombre.
- Canvas : **obligatoire** de définir `canvas.style.width/height` dans `resize()` en plus de `canvas.width/height` — sinon le canvas déborde de l'écran sur tablette/téléphone (DPR > 1).
- Sécurité : ne jamais servir de fichiers arbitraires ; conserver `validGameName` et `securityHeaders`. Le CSP autorise les scripts inline (`'unsafe-inline'`).

## Pièges connus
- L'index est généré depuis `<title>` : un jeu sans `<title>` apparaît avec son nom de fichier.
- Ne pas référencer des propriétés inexistantes sur les objets de données (ex. `e.score`) — lire depuis la table de types (`TYPES[e.type].score`), sinon `NaN`.
- Ne pas exposer les chemins internes dans les réponses HTTP : détails dans les logs uniquement.
- Les tests utilisent `server/testdata/` (copie minimale de `html/`), jamais `html/` réel.
