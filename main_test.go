package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWhoamiNeedsLogin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	res := hitGet(srv, "/api/v1/whoami")
	expectFail(t, res, http.StatusUnauthorized, "401 unauthorized")
}

func TestRootReturnsNotFound(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	res := hitGet(srv, "/")
	expectStatusCode(t, res, http.StatusNotFound)
	expectBody(t, res, "404 page not found\n")
}

func TestLogin(t *testing.T) {
	t.Parallel()
	srv := newTestServerWithUsers(t, map[string][64]byte{
		"prokop": hashPassword("catboy123"),
	})

	expectFail(t, hitPost(t, srv, "/api/v1/login", loginRequest{Username: "prokop", Password: hashPassword("eek")}), http.StatusForbidden, "wrong name or password")
	expectFail(t, hitPost(t, srv, "/api/v1/login", loginRequest{Username: "prokop", Password: hashPassword("uuhk")}), http.StatusForbidden, "wrong name or password")
	expectFail(t, hitPost(t, srv, "/api/v1/login", loginRequest{Username: "marek", Password: hashPassword("catboy123")}), http.StatusForbidden, "wrong name or password")
	res := hitPost(t, srv, "/api/v1/login", loginRequest{Username: "prokop", Password: hashPassword("catboy123")})
	expectStatusCode(t, res, http.StatusOK)
	response := decodeResponse[struct {
		Ok   bool `json:"ok"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}](t, res)
	expectEqual(t, response.Ok, true, "ok field of the response")
	expectStringLooksLikeToken(t, response.Data.Token)
}

func TestWhoami(t *testing.T) {
	t.Parallel()
	srv := newTestServerWithUsers(t, map[string][64]byte{"matúš": hashPassword("kadit")})

	token := loginHelper(t, srv, "matúš", "kadit")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/whoami", strings.NewReader(""))
	req.Header.Add("Authorization", token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	res := w.Result()

	expectStatusCode(t, res, http.StatusOK)
	expectBody(t, res, "{\"ok\":true,\"data\":{\"name\":\"matúš\"}}\n")
}
