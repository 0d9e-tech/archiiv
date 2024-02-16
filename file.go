package main

import (
	"encoding/json"
	"io"

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

func loadFilesFromRecords(records map[uuid.UUID]*Record) (files map[uuid.UUID]*File, err error) {
	var metaReader io.Reader
	for u, rec := range records {
		f := new(File)

		metaReader, err = rec.Open("meta")
		if err != nil {
			return
		}

		dec := json.NewDecoder(metaReader)
		err = dec.Decode(f)
		if err != nil {
			return
		}

		f.rec = rec
		files[u] = f
	}

	return
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
