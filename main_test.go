package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestEmptyServer(t *testing.T) {
	t.Parallel()

	type test struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int
	}

	tests := []test{
		{name: "root returns not found",
			method: http.MethodGet, path: "/", body: "", wantCode: http.StatusNotFound},
		{name: "fs upload complains about invalid uuids",
			method: http.MethodPost, path: "/api/v1/fs/upload/1/2", body: "", wantCode: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv, dir, err := newTestServer()
			if err != nil {
				t.Errorf("new test server: %v", err)
				return
			}
			defer os.RemoveAll(dir)

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
			_, err = ioutil.ReadAll(res.Body)

			if res.StatusCode != tc.wantCode {
				t.Errorf("expected status code %v. got %v", tc.wantCode, res.StatusCode)
			}
		})
	}
}
