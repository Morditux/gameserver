# ---- Étape de compilation ----
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copie du go.mod pour profiter du cache des dépendances Docker
COPY go.mod ./
RUN go mod download

# Copie du reste des sources
COPY . .

# Compilation de l'exécutable statique
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gameserver .

# ---- Étape d'exécution ----
FROM alpine:3.20

# Certificats racines (connexions HTTPS éventuelles)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copie du binaire compilé
COPY --from=builder /gameserver /app/gameserver

# Copie des fichiers statiques (html)
COPY html/ /app/html/

# Port d'écoute du serveur
EXPOSE 8080

# Exécution avec un utilisateur non-root
USER 65532:65532

CMD ["/app/gameserver"]
