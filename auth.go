package main

import (
	"archiiv/user"
	"net/http"
)

func getSessionToken(r *http.Request) string {
	return r.Header.Get("authorization")
}

func getUsername(r *http.Request) string {
	// This function is only called in endpoints wrapped around
	// `requireLogin` middleware so this function can assume that some user
	// is logged in
	// TODO(matěj)
	return ""
}

func validateToken(secret, token string) bool {
	// TODO(matěj) validate session token
	return true
}

func login(name, pwd string, userStore user.UserStore) (ok bool, token string) {
	// TODO(matěj) generate session token
	token = ""
	ok = true
	return
}
