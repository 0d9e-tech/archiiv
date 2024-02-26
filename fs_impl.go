package main

// the directory tree is modeled using the Records structs
// they are reference counted and thus are forbidden to form cycles

// Records are saved as $fs_root/$uuid

// Records contain sections saved as $fs_root/$uuid.$section The file payload
// is saved in the 'data' section. metadata is in 'meta'. hooks can create own
// sections

// External function, which take UUIDs as inputs are thread safe. Internal
// functions, which take pointers to records instead are not thread safe.

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/google/uuid"
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
	id       uuid.UUID   `json:"-"`
	refs     uint        `json:"-"`
	mutex    sync.Mutex  `json:"-"`
}

func (r *Record) lock() {
	r.mutex.Lock()
}

func (r *Record) unlock() {
	r.mutex.Unlock()
}

type Fs struct {
	lock     sync.RWMutex
	records  map[uuid.UUID]*Record
	root     uuid.UUID
	basePath string
}

func (fs *Fs) getRecord(u uuid.UUID) (*Record, error) {
	fs.lock.RLock()
	defer fs.lock.Unlock()
	r, e := fs.records[u]
	if e {
		return nil, errors.New("uuid doesn't exist")
	}
	return r, nil
}

func (fs *Fs) setRecord(r *Record) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.records[r.id] = r
}

func (fs *Fs) path(p string) string {
	// TODO sanitize paths
	return filepath.Join(fs.basePath, p)
}

func (fs *Fs) writeRecord(r *Record) error {
	f, err := os.Create(fs.path(r.id.String()))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(r)
}

func (fs *Fs) newRecord(parent *Record, name string, dir bool) (*Record, error) {
	child := new(Record)
	child.Children = []uuid.UUID{}
	child.id = uuid.New()
	child.Name = name
	child.refs = 1
	child.IsDir = dir

	fs.setRecord(child)

	parent.Children = append(parent.Children, child.id)

	return child, fs.writeRecord(child)
}

// return new slice that does not contain v
func removeUUID(s []uuid.UUID, v uuid.UUID) ([]uuid.UUID, error) {
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
			return s, errors.New("duplicite uuid")
		}
	}

	if pos == -1 {
		return s, errors.New("uuid not found")
	}

	// swap remove
	s[pos] = s[len(s)-1]
	return s[:len(s)-1], nil
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

func checkSectionNameSanity(section string) error {
	match, _ := regexp.MatchString(onlySectionPattern, section)
	if !match {
		return errors.New("section name is not sane")
	}
	return nil
}

func (fs *Fs) getSectionFileName(file uuid.UUID, section string) string {
	return fs.path(file.String() + "." + section)
}

func (fs *Fs) deleteRecord(r *Record) error {
	for _, u := range r.Children {
		err := fs.unmount(r.id, u)
		if err != nil {
			return err
		}
	}

	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		return err
	}

	idStr := r.id.String()
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), idStr) {
			os.Remove(e.Name())
		}
	}

	return nil
}

// fileStorer impl

func (fs *Fs) getChildren(u uuid.UUID) ([]uuid.UUID, error) {
	return fs.records[u].Children, nil
}

func (fs *Fs) mkdir(parentUUID uuid.UUID, name string) (uuid.UUID, error) {
	parent, err := fs.getRecord(parentUUID)
	if err != nil {
		return uuid.UUID{}, nil
	}

	r, err := fs.newRecord(parent, name, true)
	return r.id, err
}

func (fs *Fs) touch(parentUUID uuid.UUID, name string) (uuid.UUID, error) {
	parent, err := fs.getRecord(parentUUID)
	if err != nil {
		return uuid.UUID{}, err
	}

	r, err := fs.newRecord(parent, name, false)
	return r.id, err
}

func (fs *Fs) unmount(parentUUID uuid.UUID, childUUID uuid.UUID) error {
	parent, err := fs.getRecord(parentUUID)
	if err != nil {
		return err
	}

	parent.lock()
	defer parent.unlock()

	parent.Children, err = removeUUID(parent.Children, childUUID)
	if err != nil {
		return err
	}

	err = fs.writeRecord(parent)
	if err != nil {
		return err
	}

	child, err := fs.getRecord(childUUID)
	if err != nil {
		return err
	}

	child.lock()
	defer child.unlock()

	child.refs--
	if child.refs == 0 {
		return fs.deleteRecord(child)
	}

	return nil
}

func (fs *Fs) mount(parent uuid.UUID, newChild uuid.UUID) error {
	child, err := fs.getRecord(newChild)
	if err != nil {
		return err
	}

	rec, err := fs.getRecord(parent)
	if err != nil {
		return err
	}
	rec.lock()
	defer rec.unlock()

	for _, child := range rec.Children {
		if child == newChild {
			return errors.New("child with this uuid already exists")
		}
	}

	rec.Children = append(rec.Children, newChild)

	child.lock()
	child.refs++
	child.unlock()

	return fs.writeRecord(rec)
}

func (fs *Fs) openSection(uuid uuid.UUID, section string) (io.ReadCloser, error) {
	err := checkSectionNameSanity(section)
	if err != nil {
		return nil, err
	}
	return os.Open(fs.getSectionFileName(uuid, section))
}

func (fs *Fs) createSection(uuid uuid.UUID, section string) (io.WriteCloser, error) {
	err := checkSectionNameSanity(section)
	if err != nil {
		return nil, err
	}

	return os.Create(fs.getSectionFileName(uuid, section))
}

func (fs *Fs) deleteSection(uuid uuid.UUID, section string) error {
	err := checkSectionNameSanity(section)
	if err != nil {
		return err
	}

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

func newFs(root uuid.UUID, basePath string) (fs *Fs, err error) {
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
