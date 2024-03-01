package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"archiiv/fs"
)

func newTestServer() (http.Handler, string, error) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	dir, root_uuid, err := fs.InitFsDir()

	if err != nil {
		return nil, "", fmt.Errorf("init fs dir: %w", err)
	}

	srv, _, err := createServer(log, []string{
		"--fs_root", filepath.Join(dir, "fs"),
		"--users_path", filepath.Join(dir, "users.json"),
		"--root_uuid", root_uuid.String(),
	}, func(s string) string {
		if s == "ARCHIIV_SECRET" {
			return "debug_secret_321"
		}
		return ""
	})

	if err != nil {
		return nil, "", fmt.Errorf("create server: %w", err)
	}

	return srv, dir, nil
}
