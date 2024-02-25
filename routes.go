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
	mux.Handle("POST /api/v1/login", handleLogin(secret, log, userStore))
	mux.Handle("GET /api/v1/whoami", requireLogin(secret, handleWhoami()))

	mux.Handle("GET /api/v1/fs/ls/{uuid}", requireLogin(secret, handleLs(userStore, fileStore)))
	mux.Handle("GET /api/v1/fs/cat/{uuid}/{section}", requireLogin(secret, http.NotFoundHandler()))
	mux.Handle("POST /api/v1/fs/upload/{uuid}/{section}", requireLogin(secret, http.NotFoundHandler()))
	mux.Handle("POST /api/v1/fs/touch/{uuid}/{name}", requireLogin(secret, http.NotFoundHandler()))
	mux.Handle("POST /api/v1/fs/mount/{uuid}/{section}", requireLogin(secret, http.NotFoundHandler()))
	mux.Handle("POST /api/v1/fs/unmount/{uuid}/{section}", requireLogin(secret, http.NotFoundHandler()))

	mux.Handle("/", http.NotFoundHandler())
}
