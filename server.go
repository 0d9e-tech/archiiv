package main

import (
	"log/slog"
	"net/http"
)

func newServer(
	logger *slog.Logger,
	secret string,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, secret)
	var handler http.Handler = mux
	return handler
}
