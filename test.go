package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// /tmp/archiiv_test_295721899
// ├── fs
// │   ├── ...
// │   ├── ...
// │   └── 38b4183d-4df4-43dd-9495-1847083a3662
// └── users.json
func initFsDir() (dir string, root_uuid uuid.UUID, err error) {
	dir, err = os.MkdirTemp("", "archiiv_test_*")
	if err != nil {
		return
	}

	fs_dir := filepath.Join(dir, "fs")
	if err = os.Mkdir(fs_dir, 0755); err != nil {
		return
	}

	root_uuid = uuid.New()

	// create root uuid
	f, err := os.Create(filepath.Join(fs_dir, root_uuid.String()))
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(Record{
		IsDir: true,
	})

	// create users.json
	users_path := filepath.Join(dir, "users.json")
	f2, err := os.Create(users_path)
	if err != nil {
		return
	}
	defer f2.Close()

	err = json.NewEncoder(f2).Encode(map[string]string{})

	return
}

func newTestServer() (http.Handler, string, error) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	dir, root_uuid, err := initFsDir()

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
