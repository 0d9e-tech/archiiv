package main

import (
	"log/slog"
	"net/http"
)

func logAccesses(logger *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "url", r.URL.Path)
		h.ServeHTTP(w, r)
	})
}
