package fs

import (
	"encoding/json"
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
func InitFsDir() (dir string, rootUUID uuid.UUID, err error) {
	dir, err = os.MkdirTemp("", "archiiv_test_*")
	if err != nil {
		return
	}

	fsDir := filepath.Join(dir, "fs")
	if err = os.Mkdir(fsDir, 0755); err != nil {
		return
	}

	rootUUID = uuid.New()

	// create root uuid
	f, err := os.Create(filepath.Join(fsDir, rootUUID.String()))
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(record{
		IsDir: true,
	})

	// create users.json
	usersPath := filepath.Join(dir, "users.json")
	f2, err := os.Create(usersPath)
	if err != nil {
		return
	}
	defer f2.Close()

	err = json.NewEncoder(f2).Encode(map[string]string{})

	return
}
