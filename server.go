package main

import (
	"log/slog"
	"net/http"
)

func NewServer(
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger)
	var handler http.Handler = mux
	return handler
}
