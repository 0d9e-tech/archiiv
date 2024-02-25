package main

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

const (
	PermOwner = uint8(1 << iota)
	PermRead
	PermWrite
)

// Metadata stored with each file in it's 'meta' section
type fileMeta struct {
	UUID      uuid.UUID        `json:"uuid"`
	Type      string           `json:"type"`
	Perms     map[string]uint8 `json:"perms"`
	Hooks     []string         `json:"hooks"`
	CreatedBy string           `json:"createdBy"`
	CreatedAt uint64           `json:"createdAt"`
	rec       *Record
}

func readFileMeta(fs fileStorer, file uuid.UUID) (fm fileMeta, err error) {
	data, err := fs.readSection(file, "meta")
	if err != nil {
		return
	}

	err = json.NewDecoder(strings.NewReader(string(data))).Decode(&fm)
	return
}

func writeFileMeta(fs fileStorer, file uuid.UUID, fm fileMeta) error {
	encoded, err := json.Marshal(fm)
	if err != nil {
		return err
	}

	return fs.writeSection(file, "meta", encoded)
}
