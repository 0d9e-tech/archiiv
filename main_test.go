package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type responseError struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func decodeResponse[T any](t *testing.T, r *http.Response) (v T) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&v)

	if err != nil {
		t.Errorf("failed to decode response %v", err)
	} else if _, err := dec.Token(); err != io.EOF { // check that there is nothing leaking after the json
		t.Error("json contains trailing data")
	}

	return
}

func expectEqual[T comparable](t *testing.T, got, expected T, comment string) {
	if expected != got {
		t.Errorf("%s should be %#v (is %#v)", comment, expected, got)
	}
}

func expectStatusCode(t *testing.T, res *http.Response, expected int) {
	expectEqual(t, res.StatusCode, expected, "status code")
}

func expectBody(t *testing.T, res *http.Response, expected string) {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read response body: %v", err)
	}
	expectEqual(t, string(b), expected, "response body")
}

func expectFail(t *testing.T, res *http.Response, statusCode int, errorMessage string) {
	expectStatusCode(t, res, statusCode)
	b := decodeResponse[responseError](t, res)
	expectEqual(t, b.Ok, false, "ok field")
	expectEqual(t, b.Error, errorMessage, "response body")
}

func hit(srv http.Handler, method, target string, body io.Reader) *http.Response {
	req := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w.Result()
}

func hitPost(t *testing.T, srv http.Handler, target string, body any) *http.Response {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		t.Errorf("failed to encode post body: %v", err)
	}

	return hit(srv, http.MethodPost, "/api/v1/login", &buf)
}

func hitGet(srv http.Handler, target string) *http.Response {
	req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w.Result()
}

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
		"prokop": HashPassword("catboy123"),
	})

	type LoginRequest struct {
		Username string   `json:"username"`
		Password [64]byte `json:"password"`
	}

	expectFail(t, hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "prokop", Password: HashPassword("eek")}),
		http.StatusForbidden, "wrong name or password")
	expectFail(t, hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "prokop", Password: HashPassword("uuhk")}),
		http.StatusForbidden, "wrong name or password")
	expectFail(t, hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "marek", Password: HashPassword("catboy123")}),
		http.StatusForbidden, "wrong name or password")
	res := hitPost(t, srv, "/api/v1/login", LoginRequest{Username: "prokop", Password: HashPassword("catboy123")})
	expectStatusCode(t, res, http.StatusOK)
	expectBody(t, res, "{\"ok\":true,\"data\":{\"token\":\"Ff+DAwEBCUZ1bGxUb2tlbgH/hAAAAA==\"}}\n")
}

func TestFsUploadUUIDParse(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	// TODO login here
	res := hit(srv, http.MethodPost, "/api/v1/fs/upload/1/2", strings.NewReader(""))
	expectFail(t, res, http.StatusBadRequest, "invalid uuid")
}
