package main

import (
	"log/slog"
	"net/http"
)

func logAccesses(log *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", "url", r.URL.Path)
		h.ServeHTTP(w, r)
	})
}
