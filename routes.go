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

	mux.Handle("GET /api/v1/fs/ls/{uuid}", requireLogin(secret, handleLs(fileStore)))
	mux.Handle("GET /api/v1/fs/cat/{uuid}/{section}", requireLogin(secret, handleCat(fileStore)))
	mux.Handle("POST /api/v1/fs/upload/{uuid}/{section}", requireLogin(secret, handleUpload(fileStore)))
	mux.Handle("POST /api/v1/fs/touch/{uuid}/{name}", requireLogin(secret, handleTouch(fileStore)))
	mux.Handle("POST /api/v1/fs/mkdir/{uuid}/{name}", requireLogin(secret, handleMkdir(fileStore)))
	mux.Handle("POST /api/v1/fs/mount/{parentUuid}/{childUuid}", requireLogin(secret, handleMount(fileStore)))
	mux.Handle("POST /api/v1/fs/unmount/{parentUuid}/{childUuid}", requireLogin(secret, handleUnmount(fileStore)))

	mux.Handle("/", http.NotFoundHandler())
}
