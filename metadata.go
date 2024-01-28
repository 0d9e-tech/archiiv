package main

import (
	"encoding/json"
	"os"
)

type FileEdit struct {
	User string
	Date uint64
	Note string
}

type Metadata struct {
	Perms     Perms                  `json:"perms"`
	Hooks     []string               `json:"hooks"`
	CreatedBy string                 `json:"createdBy"`
	Edits     []FileEdit             `json:"edits"`
	HookData  map[string]interface{} `json:"hookData"`
}

func getMetadata(file string) (Metadata, error) {
	m := Metadata{}

	f, err := os.Open(toMetadataPath(file))
	if err != nil {
		return m, err
	}

	dec := json.NewDecoder(f)
	err = dec.Decode(&m)
	if err != nil {
		return m, err
	}

	return m, nil
}

func (m *Metadata) write(file string) error {
	f, err := os.Create(toMetadataPath(file))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(m)
}
