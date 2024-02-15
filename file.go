package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

const (
	PermOwner = uint8(1 << iota)
	PermRead
	PermWrite
)

type File struct {
	UUID      uuid.UUID        `json:"uuid"`
	Type      string           `json:"type"`
	Perms     map[string]uint8 `json:"perms"`
	Hooks     []string         `json:"hooks"`
	CreatedBy string           `json:"createdBy"`
	CreatedAt uint64           `json:"createdAt"`
	rec       *Record
}

func (av *Archiiv) loadFiles() error {
	av.files = map[uuid.UUID]*File{}

	for u, rec := range av.fs.records {
		f := new(File)

		metaReader, err := rec.Open("meta")
		if err != nil {
			return err
		}

		dec := json.NewDecoder(metaReader)
		err = dec.Decode(f)
		if err != nil {
			return err
		}

		f.rec = rec
		av.files[u] = f
	}

	return nil
}

func (f *File) Save() error {
	w, err := f.rec.Create("meta")
	if err != nil {
		return err
	}
	defer w.Close()

	enc := json.NewEncoder(w)
	return enc.Encode(f)
}
