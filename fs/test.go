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
func InitFsDir() (dir string, root_uuid uuid.UUID, err error) {
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

	err = json.NewEncoder(f).Encode(record{
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
