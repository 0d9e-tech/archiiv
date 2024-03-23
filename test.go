package main

import (
	"archiiv/fs"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) http.Handler {
	return newTestServerWithUsers(t, map[string][64]byte{})
}

func newTestServerWithUsers(t *testing.T, users map[string][64]byte) http.Handler {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	dir, rootUUID := fs.InitFsDir(t, users)

	secret := generateSecret()

	srv, _, err := createServer(log, []string{
		"--fs_root", filepath.Join(dir, "fs"),
		"--users_path", filepath.Join(dir, "users.json"),
		"--root_uuid", rootUUID.String(),
	}, func(s string) string {
		if s == "ARCHIIV_SECRET" {
			return secret
		}
		return ""
	})

	if err != nil {
		t.Fatalf("newTestServer: %v", err)
	}

	return srv
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

func expectStringLooksLikeToken(t *testing.T, token string) {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Errorf("string does not look like token: base64 decode: %v", err)
	}

	ft, err := gobDecode[fullToken](data)
	if err != nil {
		t.Errorf("string does not look like token: gob decode %v", err)
	}

	_, err = payloadToBytes(ft.Data)
	if err != nil {
		t.Errorf("string does not look like token: payload to bytes: %v", err)
	}
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

type loginRequest struct {
	Username string   `json:"username"`
	Password [64]byte `json:"password"`
}

func loginHelper(t *testing.T, srv http.Handler, username, pwd string) string {
	res := hitPost(t, srv, "/api/v1/login", loginRequest{Username: username, Password: hashPassword(pwd)})

	type LoginResponse struct {
		Ok   bool `json:"ok"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	lr := decodeResponse[LoginResponse](t, res)

	return lr.Data.Token
}
