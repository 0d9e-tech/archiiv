package main

import (
	"log/slog"
	"net/http"
)

func newServer(
	logger *slog.Logger,
	secret string,
	userStore userStorer,
	fileStore fileStorer,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		logger,
		secret,
		userStore,
		fileStore,
	)
	var handler http.Handler = mux
	return handler
}
