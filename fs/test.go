package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// InitFsDir creates the following directory structure:
//
// /tmp/archiiv_test_295721899
// ├── fs
// │   ├── ...
// │   ├── ...
// │   └── 38b4183d-4df4-43dd-9495-1847083a3662
// └── users.json
func InitFsDir(t *testing.T) (dir string, rootUUID uuid.UUID) {
	dir = t.TempDir()

	fsDir := filepath.Join(dir, "fs")
	if err := os.Mkdir(fsDir, 0750); err != nil {
		t.Fatalf("InitFsDir: %v", err)
	}

	rootUUID = uuid.New()

	// create root uuid
	f, err := os.Create(filepath.Join(fsDir, rootUUID.String()))
	if err != nil {
		t.Fatalf("InitFsDir: %v", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(record{
		IsDir: true,
	})
	if err != nil {
		t.Fatalf("InitFsDir: %v", err)
	}

	// create users.json
	usersPath := filepath.Join(dir, "users.json")
	f2, err := os.Create(usersPath)
	if err != nil {
		t.Fatalf("InitFsDir: %v", err)
	}
	defer f2.Close()

	err = json.NewEncoder(f2).Encode(map[string]string{})
	if err != nil {
		t.Fatalf("InitFsDir: %v", err)
	}

	return
}
