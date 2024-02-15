package main

import (
	"log/slog"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	secret string,
	userStore userStorer,
	fileStore fileStorer,
) {
	mux.Handle("/api/v1/login", handleLogin(secret, logger, userStore))
	mux.Handle("/api/v1/fs/ls", requireLogin(secret, handleLs(logger, userStore, fileStore)))
	mux.Handle("/", http.NotFoundHandler())
}
