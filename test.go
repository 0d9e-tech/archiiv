package main

import (
	"archiiv/fs"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"
)

func newTestServer(t *testing.T) http.Handler {
	return newTestServerWithUsers(t, map[string][64]byte{})
}

func newTestServerWithUsers(t *testing.T, users map[string][64]byte) http.Handler {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	dir, rootUUID := fs.InitFsDir(t, users)

	secret := GenerateSecret()

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
