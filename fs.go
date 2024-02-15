package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// File/Dir
type Record struct {
	Name     string      `json:"name"`
	Children []uuid.UUID `json:"children"`
	UUID     uuid.UUID   `json:"-"`
	refs     int
	fs       *Fs
}

type fileStorer interface {
	get(uuid uuid.UUID) Record
}

// NOTE(mrms): The remove calls in this function may fail, but in reality, it's
// higly improbable.
func (rec *Record) delete() {
	name := rec.UUID.String()
	os.Remove(rec.fs.path(name))

	entries, _ := os.ReadDir(rec.fs.path(""))
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), name+".") {
			continue
		}
		os.Remove(rec.fs.path(e.Name()))
	}

	delete(rec.fs.records, rec.UUID)
}

func (rec *Record) incRef() {
	for _, c := range rec.Children {
		rec.fs.records[c].incRef()
	}

	rec.refs++
}

func (rec *Record) decRef() {
	for _, c := range rec.Children {
		rec.fs.records[c].incRef()
	}

	rec.refs--
	if rec.refs == 0 {
		rec.delete()
	}
}

func (rec *Record) Save() error {
	f, err := os.Create(rec.fs.path(rec.UUID.String()))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(*rec)
}

// TODO: Check if there isn't a record with the same name in the folder already.
func (parent *Record) Mount(rec *Record) {
	parent.Children = append(parent.Children, rec.UUID)
	parent.Save()
	rec.incRef()
}

func (parent *Record) Unmount(rec *Record) {
	for i, u := range parent.Children {
		if u == rec.UUID {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
		}
	}

	parent.Save()
	rec.decRef()
}

// TODO: Sanitize that the section name doesn't contain slashes or other
// nasty characters.
func (rec *Record) Open(section string) (*os.File, error) {
	return os.Open(rec.fs.path(rec.UUID.String() + "." + section))
}

func (rec *Record) Create(section string) (*os.File, error) {
	return os.Create(rec.fs.path(rec.UUID.String() + "." + section))
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

func (fs *Fs) loadRecords() error {
	entries, err := os.ReadDir(fs.path(""))
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.Type().IsDir() {
			continue
		}

		name := e.Name()

		u, err := uuid.Parse(name)
		if err != nil { // Not a valid UUID
			return err
		}

		f, err := os.Open(fs.path(name))
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

		rec.UUID = u
		rec.fs = fs
		fs.records[u] = rec
	}

	return err
}

func (fs *Fs) runGC() error {
	fs.records[fs.root].incRef()

	return nil
}

func (fs *Fs) NewRecord(name string) (*Record, error) {
	rec := new(Record)
	rec.fs = fs
	rec.UUID = uuid.New()
	rec.Name = name
	rec.Children = []uuid.UUID{}

	fs.records[rec.UUID] = rec

	return rec, rec.Save()
}

func (fs *Fs) GetRecord(u uuid.UUID) *Record {
	return fs.records[u]
}

func NewFs(root uuid.UUID, basePath string) (fs Fs, err error) {
	fs.basePath = basePath
	fs.root = root
	fs.records = make(map[uuid.UUID]*Record)

	err = fs.loadRecords()
	if err != nil {
		return
	}

	if _, c := fs.records[root]; !c {
		return fs, errors.New("the root UUID not found in fs")
	}

	// TODO: Maybe we should check for cycles here?

	err = fs.runGC()
	if err != nil {
		return
	}

	return
}

func handleLs(logger *slog.Logger, userStore userStorer, fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := getUser(r)
		uuid, err := uuid.Parse(r.FormValue("uuid"))
		if err != nil {
			encodeError(w, http.StatusBadRequest, "invalid uuid")
		}

		record := fileStore.get(uuid)

		if !hasReadPerm(username, record) {
			encodeError(w, http.StatusForbidden, "Insufficient permissions")
			return
		}

		encodeOK(w, http.StatusOK, record.Children)
	})
}
