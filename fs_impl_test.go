package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmptyServer(t *testing.T) {
	t.Parallel()
	srv, err := newTestServer()
	if err != nil {
		t.Errorf("new test server: %v", err)
		return
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fs/cat/1/2", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)

	if (res.StatusCode != 400) {
		t.Errorf("expected to fail. got %v", res.StatusCode)
	}
}
