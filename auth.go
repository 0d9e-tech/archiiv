package main

import (
	"log/slog"
	"net/http"
)

func getSessionToken(r *http.Request) string {
	return r.Header.Get("authorization")
}

func getUser(r *http.Request) string {
	// TODO
	return ""
}

func validateToken(secret, token string) bool {
	// TODO validate session token
	return true
}

func login(name, pwd string, userStore userStorer) (ok bool, token string) {
	// TODO generate session token
	token = ""
	return
}

func handleLogin(secret string, logger *slog.Logger, userStore userStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		pwd := r.FormValue("pwd")

		ok, token := login(name, pwd, userStore)

		if ok {
			logger.Info("New login", "user", name)
			encodeOK(w, http.StatusOK, struct {
				Token string `json:"token"`
			}{Token: token})
		} else {
			logger.Info("Failed login", "user", name)
			encodeError(w, http.StatusForbidden, "Wrong name or password")
		}
	})
}

func requireLogin(secret string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getSessionToken(r)
		if validateToken(secret, token) {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, "401 unauthorized", http.StatusUnauthorized)
		}
	})
}
