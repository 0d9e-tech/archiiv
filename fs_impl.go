package main

// the directory tree is modeled using the Records structs
// they are reference counted and thus are forbidden to form cycles

// Records are saved as $fs_root/$uuid

// Records contain sections saved as $fs_root/$uuid.$section The file payload
// is saved in the 'data' section. metadata is in 'meta'. hooks can create own
// sections

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	sectionPattern      = `[a-zA-Z0-9_-]+`
	uuidPattern         = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	fileInFsRootPattern = uuidPattern + `(\.` + sectionPattern + `)?`

	onlyUuidPattern         = `^` + uuidPattern + `$`
	onlyFileInFsRootPattern = `^` + fileInFsRootPattern + `$`
	onlySectionPattern      = `^` + sectionPattern + `$`
)

type Record struct {
	Children []uuid.UUID `json:"children,omitempty"`
	IsDir    bool        `json:"is_dir"`
	Name     string      `json:"name"`
	refs     uint        `json:"-"`
}

type Fs struct {
	records  map[uuid.UUID]*Record
	root     uuid.UUID
	basePath string
}

func (fs *Fs) path(p string) string {
	// TODO sanitize paths
	return filepath.Join(fs.basePath, p)
}

func (fs *Fs) writeToDisk(id uuid.UUID, r *Record) error {
	f, err := os.Create(fs.path(id.String()))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(r)
}

func (fs *Fs) newRecord(parent_uuid uuid.UUID, name string) (uuid.UUID, error) {
	new_id := uuid.New()

	new_rec := new(Record)
	new_rec.Children = []uuid.UUID{}
	new_rec.Name = name
	new_rec.refs = 1

	fs.records[new_id] = new_rec

	parent_record := fs.records[parent_uuid]
	parent_record.Children = append(parent_record.Children, new_id)
	fs.records[parent_uuid] = parent_record

	return new_id, fs.writeToDisk(new_id, new_rec)
}

// return new slice that does not contain v
func removeUUID(s []uuid.UUID, v uuid.UUID) []uuid.UUID {
	i := 0
	pos := -1
	for ; i < len(s); i++ {
		if s[i] == v {
			pos = i
			break
		}
	}

	for ; i < len(s); i++ {
		if s[i] == v {
			panic("duplicite uuid")
		}
	}

	if pos == -1 {
		panic("uuid not found")
	}

	// swap remove
	s[pos] = s[len(s)-1]
	return s[:len(s)-1]
}

func (fs *Fs) deleteRecordFilesFromDisk(name string) error {
	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		return err
	}

	// list of all errors. We try to delete as many files as we can to
	// avoid zombie files and then return one big error with all of the
	// failed removes.
	var errs []error

	err = os.Remove(fs.path(name))
	if err != nil {
		errs = append(errs, err)
	}

	// TODO(prokop) keep a list of all sections instead of iterating over
	// all files
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), name+".") {
			err = os.Remove(fs.path(e.Name()))
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (fs *Fs) deleteChildInMemoryAndWriteParentToDisk(parent uuid.UUID, child uuid.UUID) error {
	ch := fs.records[parent].Children
	fs.records[parent].Children = removeUUID(ch, child)

	return fs.writeToDisk(parent, fs.records[parent])
}

func (fs *Fs) decreseChildrenRefCountRecursive(current uuid.UUID) error {
	// list of all errors. We try to delete as many files as we can to
	// avoid zombie files and then return one big error with all of the
	// failed removes.
	var errs []error

	for _, child := range fs.records[current].Children {
		if fs.records[child].refs == 0 {
			panic("decresing ref count below zero")
		}

		fs.records[child].refs--

		if fs.records[child].refs == 0 {
			err := fs.delete(current, child)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

func checkSectionNameIsSane(section string) error {
	match, _ := regexp.MatchString(onlySectionPattern, section)
	if !match {
		return errors.New("section name is not sane")
	}
	return nil
}

func (fs *Fs) getSectionFileName(file uuid.UUID, section string) string {
	return fs.path(file.String() + "." + section)
}

// fileStorer impl

func (fs *Fs) getChildren(u uuid.UUID) ([]uuid.UUID, error) {
	return fs.records[u].Children, nil
}

func (fs *Fs) mkdir(parent_uuid uuid.UUID, name string) (uuid.UUID, error) {
	return fs.newRecord(parent_uuid, name)
}

func (fs *Fs) touch(parent_uuid uuid.UUID, name string) (uuid.UUID, error) {
	return fs.newRecord(parent_uuid, name)
}

func (fs *Fs) delete(parent uuid.UUID, child uuid.UUID) error {
	err := fs.deleteChildInMemoryAndWriteParentToDisk(parent, child)
	if err != nil {
		return err
	}

	err = fs.deleteRecordFilesFromDisk(child.String())
	if err != nil {
		return err
	}

	return fs.decreseChildrenRefCountRecursive(parent)
}

func (fs *Fs) mount(parent uuid.UUID, newChild uuid.UUID) error {
	rec := fs.records[parent]

	for _, child := range rec.Children {
		if child == newChild {
			return errors.New("child with this uuid already exists")
		}
	}

	rec.Children = append(rec.Children, newChild)

	fs.records[newChild].refs++

	return fs.writeToDisk(parent, rec)
}

func (fs *Fs) readSection(uuid uuid.UUID, section string) ([]byte, error) {
	err := checkSectionNameIsSane(section)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(fs.getSectionFileName(uuid, section))
}

func (fs *Fs) writeSection(uuid uuid.UUID, section string, data []byte) error {
	err := checkSectionNameIsSane(section)
	if err != nil {
		return err
	}

	return os.WriteFile(fs.getSectionFileName(uuid, section), data, 0644)
}

func (fs *Fs) deleteSection(uuid uuid.UUID, section string) error {
	return os.Remove(fs.getSectionFileName(uuid, section))
}

//---

func (fs *Fs) loadRecords() error {
	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		return err
	}

	var sectionFiles []string
	var recordFiles []string

	for _, e := range entries {
		if e.Type().IsDir() {
			return errors.New("garbage directory in fs root")
		}

		name := e.Name()

		match, _ := regexp.MatchString(onlyFileInFsRootPattern, name)
		if !match {
			return errors.New("garbage file in fs root")
		}

		if len(name) == 36 {
			recordFiles = append(recordFiles, name)
		} else {
			sectionFiles = append(sectionFiles, name)
		}
	}

	for _, recordName := range recordFiles {
		u, err := uuid.Parse(recordName)
		if err != nil {
			return err
		}

		f, err := os.Open(fs.path(recordName))
		if err != nil {
			return err
		}
		defer f.Close()

		rec := new(Record)
		dec := json.NewDecoder(f)
		err = dec.Decode(rec)
		if err != nil {
			return err
		}

		fs.records[u] = rec
	}

	// TODO(prokop) load section file names
	return nil
}

func checkLoadedRecordsAreSane(map[uuid.UUID]*Record) error {
	// TODO(prokop)
	return nil
}

func newFs(root uuid.UUID, basePath string) (fs Fs, err error) {
	fs.basePath = basePath
	fs.root = root
	fs.records = make(map[uuid.UUID]*Record)

	err = fs.loadRecords()
	if err != nil {
		return
	}

	if _, c := fs.records[root]; !c {
		err = errors.New("the root UUID not found in fs")
		return
	}

	return fs, checkLoadedRecordsAreSane(fs.records)
}

func handleLs(userStore userStorer, fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		username := getUsername(r)

		uuid, err := uuid.Parse(uuidArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, "invalid uuid")
			return
		}

		children, err := fileStore.getChildren(uuid)
		if err != nil {
			encodeError(w, http.StatusNotFound, "file not found")
			return
		}

		if !hasReadPerm(username, uuid) {
			encodeError(w, http.StatusForbidden, "insufficient permissions")
			return
		}

		encodeOK(w, children)
	})
}
