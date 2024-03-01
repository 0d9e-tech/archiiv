package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func getSessionToken(r *http.Request) string {
	return r.Header.Get("authorization")
}

func getUsername(r *http.Request) string {
	// This function is only called in endpoints wrapped around
	// `requireLogin` middleware so this function can assume that some user
	// is logged in
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
	ok = true
	return
}

func handleLogin(secret string, log *slog.Logger, userStore userStorer) http.Handler {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)

		var lr LoginRequest

		if err := dec.Decode(&lr); err != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("invalid json: %w", err))
			return
		}

		name := r.FormValue("name")
		pwd := r.FormValue("pwd")

		ok, token := login(name, pwd, userStore)

		if ok {
			log.Info("New login", "user", name)
			encodeOK(w, map[string]any{"token": token})
			return
		} else {
			log.Info("Failed login", "user", name)
			encodeError(w, http.StatusForbidden, errors.New("wrong name or password"))
			return
		}
	})
}

func handleWhoami() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := getUsername(r)
		encodeOK(w, struct {
			Name string `json:"name"`
		}{Name: name})
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
