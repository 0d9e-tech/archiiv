package main

import (
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
	expectFail(t, res, http.StatusUnauthorized, "Unauthorized")
}

func TestRootReturnsNotFound(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	res := hitGet(srv, "/")
	expectStatusCode(t, res, http.StatusNotFound)
	expectBody(t, res, "404 page not found\n")
}

func TestFsUploadUUIDParse(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	// TODO login here
	res := hit(srv, http.MethodPost, "/api/v1/fs/upload/1/2", strings.NewReader(""))
	expectFail(t, res, http.StatusBadRequest, "invalid uuid")
}
