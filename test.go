package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"fmt"
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
func initFsDir() (fs_dir string, users_path string, root_uuid uuid.UUID, err error) {
	dname, err := os.MkdirTemp("", "archiiv_test_*")
	if err != nil {
		return
	}

	fs_dir = filepath.Join(dname, "fs")
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
	users_path = filepath.Join(dname, "users.json")
	f2, err := os.Create(users_path)
	if err != nil {
		return
	}
	defer f2.Close()

	err = json.NewEncoder(f2).Encode(map[string]string{})

	return
}

func newTestServer() (http.Handler, error) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	fs_root, users_path, root_uuid, err := initFsDir()
	if err != nil {
		return nil, fmt.Errorf("init fs dir: %w", err)
	}

	srv, _, err := createServer(log, []string{
		"--fs_root", fs_root,
		"--users_path", users_path,
		"--root_uuid", root_uuid.String(),
	}, func(s string) string {
		if s == "ARCHIIV_SECRET" {
			return "debug_secret_321"
		}
		return ""
	})

	if err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}

	return srv, nil
}
