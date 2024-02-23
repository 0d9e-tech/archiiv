package main

import (
	"log/slog"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	log *slog.Logger,
	secret string,
	userStore userStorer,
	fileStore fileStorer,
) {
	mux.Handle("/api/v1/login", handleLogin(secret, log, userStore))
	mux.Handle("/api/v1/whoami", requireLogin(secret, handleWhoami(log)))
	mux.Handle("/api/v1/fs/ls/{uuid}", logAccesses(log, requireLogin(secret, handleLs(userStore, fileStore))))
	mux.Handle("/", logAccesses(log, http.NotFoundHandler()))
}
