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

type LoginRequest struct {
	Username string   `json:"username"`
	Password [64]byte `json:"password"`
}

func TestLogin(t *testing.T) {
	t.Parallel()
	srv := newTestServerWithUsers(t, map[string][64]byte{
		"prokop": HashPassword("catboy123"),
	})

	expectFail(t, hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "prokop", Password: HashPassword("eek")}), http.StatusForbidden, "wrong name or password")
	expectFail(t, hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "prokop", Password: HashPassword("uuhk")}), http.StatusForbidden, "wrong name or password")
	expectFail(t, hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "marek", Password: HashPassword("catboy123")}), http.StatusForbidden, "wrong name or password")
	res := hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "prokop", Password: HashPassword("catboy123")})
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

func loginHelper(t *testing.T, srv http.Handler, username, pwd string) string {
	res := hitPost(t, srv, "/api/v1/login", LoginRequest{Username: username, Password: HashPassword(pwd)})

	type LoginResponse struct {
		Ok   bool `json:"ok"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	lr := decodeResponse[LoginResponse](t, res)

	return lr.Data.Token
}

func TestWhoamiWorks(t *testing.T) {
	t.Parallel()
	srv := newTestServerWithUsers(t, map[string][64]byte{"matúš": HashPassword("kadit")})

	token := loginHelper(t, srv, "matúš", "kadit")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/whoami", strings.NewReader(""))
	req.Header.Add("Authorization", token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	res := w.Result()

	expectStatusCode(t, res, http.StatusOK)
	expectBody(t, res, "{\"ok\":true,\"data\":{\"name\":\"matúš\"}}\n")
}
